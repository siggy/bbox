package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox/leds"
)

// LEDS
frontLeds = []int{
	0,
	29,
	55,
	79,
	102,
	124,
	145,
	165,
	183,
	200,
	216,
	230,
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	level := make(chan float64)

	baux := leds.InitBaux(level)

	go baux.Run()

	defer baux.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
