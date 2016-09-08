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
	Temperature float32 `json:"temperature"`
	Pressure    float32 `json:"pressure"`
	Altitude    float32 `json:"altitude"`
}

// Config ...
type Config struct {
	ReportingInterval int    `toml:"reporting_interval"`
	ThingName         string `toml:"thing_name"`
	ThingEndpoint     string `toml:"thing_endpoint"`
	ThingRegion       string `toml:"thing_region"`
}

var config Config

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

	for {
		pressure, err := sensor.Pressure()

		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Pressure %v", pressure)
		}

		altitude, err := sensor.Altitude()

		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Altitude %v", altitude)
		}

		temperature, err := sensor.Temperature()

		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Temperature %v", temperature)
		}

		time.Sleep(time.Second * time.Duration(config.ReportingInterval))
	}
}
