package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
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
	bus := embd.NewI2CBus(1)
	sensor := bmp180.New(bus)

	sensor.Run()
	defer sensor.Close()

	firstLoop := true

	for {
		var err error
		reading := SensorReading{}

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

		time.Sleep(time.Second * time.Duration(config.ReportingInterval))

		log.Printf("Reading: %v\n", reading)

		if firstLoop {
			firstLoop = false
			log.Print("Waiting for the sensor to stabilise...")

			// Make sure the sensor has some time to get a grip (regardless of
			// the ReportingInterval value)
			time.Sleep(time.Second * 5)
			log.Println("Done.")
		} else {
			// TODO: post update
		}
	}
}
