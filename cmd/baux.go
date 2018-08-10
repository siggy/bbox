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

	amplitude := bbox.InitAmplitude(level)
	baux := leds.InitBaux(level, web.InitWeb())

	go baux.Run()
	go amplitude.Run()

	defer baux.Close()
	defer amplitude.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
