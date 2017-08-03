package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/leds"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	level := make(chan float64)

	amplitude := bbox.InitAmplitude(level)
	crane := leds.InitCrane(level)

	go amplitude.Run()
	go crane.Run()

	defer amplitude.Close()
	defer crane.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
