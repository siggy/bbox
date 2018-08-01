package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/bbox/beatboxer/render/web"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	level := make(chan float64)
	press := make(chan struct{})

	web.InitWeb()

	amplitude := bbox.InitAmplitude(level)
	gpio := bbox.InitGpio(press)
	fish := leds.InitFish(level, press)

	go amplitude.Run()
	go gpio.Run()
	go fish.Run()

	defer amplitude.Close()
	defer gpio.Close()
	defer fish.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
