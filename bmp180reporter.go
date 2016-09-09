package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iotdataplane"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/kidoman/embd/sensor/bmp180"
)

// SensorReading ...
type SensorReading struct {
	Pressure    int     `json:"pressure"`
	Temperature float64 `json:"temperature"`
}

// Config ...
type Config struct {
	ReportingInterval int    `toml:"reporting_interval"`
	ThingName         string `toml:"thing_name"`
	ThingEndpoint     string `toml:"thing_endpoint"`
	ThingRegion       string `toml:"thing_region"`
}

var config Config

const discardCount int64 = 5

func init() {
	var help bool
	var configPath string

	flag.StringVar(&configPath, "config", "", "path to the config file")
	flag.BoolVar(&help, "help", false, "this help mesage")
	flag.Parse()

	if configPath == "" || help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	tomlData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = toml.Decode(string(tomlData), &config)
	if err != nil {
		log.Fatalln(err)
	}

	if config.ReportingInterval <= 0 {
		log.Fatalln("ReportingInterval must be greater than zero")
	}

	log.Printf("ThingName: %v, ThingEndpoint: %v, ThingRegion %v", config.ThingName, config.ThingEndpoint, config.ThingRegion)
}

func main() {
	log.Println("Initializing the I2C bus")
	bus := embd.NewI2CBus(1)

	log.Println("Initializing the BMP180 sensor")
	sensor := bmp180.New(bus)

	log.Println("Starting the BMP180 sensor")
	sensor.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			sensor.Close()
			bus.Close()
			os.Exit(1)
		}
	}()

	firstLoop := true

	for {
		var err error
		reading := SensorReading{}

		log.Println("Getting reading...")
		reading.Pressure, err = sensor.Pressure()
		if err != nil {
			log.Println(err)
			continue
		}

		reading.Temperature, err = sensor.Temperature()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Reading: %v\n", reading)

		if firstLoop {
			firstLoop = false
			log.Print("Waiting for the sensor to stabilise...")

			// Make sure the sensor has some time to get a grip (regardless of
			// the ReportingInterval value)
			time.Sleep(time.Second * 5)
			log.Println("Done.")
		} else {
			go func() {
				updateThingShadow(&reading)
			}()
			time.Sleep(time.Second * time.Duration(config.ReportingInterval))
		}
	}
}

func updateThingShadow(reading *SensorReading) {
	if config.ThingName == "" || config.ThingEndpoint == "" || config.ThingRegion == "" {
		log.Println("Missing thing_name, thing_endpoint, or thing_region configuration")
		return
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Println("Failed to create AWS session: ", err)
		return
	}

	awsConfig := aws.NewConfig().WithEndpoint(config.ThingEndpoint).WithRegion(config.ThingRegion)
	svc := iotdataplane.New(sess, awsConfig)

	payload, err := json.Marshal(map[string]map[string]SensorReading{
		"state": map[string]SensorReading{
			"reported": *reading,
		},
	})
	if err != nil {
		log.Println("Serialization error: ", err)
		return
	}

	log.Println("Updating Thing Shadow…")
	params := &iotdataplane.UpdateThingShadowInput{
		Payload:   payload,
		ThingName: aws.String(config.ThingName),
	}
	_, err = svc.UpdateThingShadow(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Println(err.Error())
		return
	}

	log.Println("Thing Shadow updated")
}
