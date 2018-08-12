package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/leds/human"
	"github.com/siggy/bbox/beatboxer/render/web"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	level := make(chan float64)

	amplitude := bbox.InitAmplitude(level)
	human := human.Init(level, web.InitWeb())

	go human.Run()
	go amplitude.Run()

	defer human.Close()
	defer amplitude.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
