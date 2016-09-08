package main

import (
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/kidoman/embd/sensor/bmp180"
)

func init() {}

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

		time.Sleep(5 * time.Second)
	}
}
