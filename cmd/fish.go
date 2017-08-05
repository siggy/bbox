package main

import (
	"sync"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/leds"
)

func main() {
	var wg sync.WaitGroup

	level := make(chan float64)
	press := make(chan struct{})

	amplitude := bbox.InitAmplitude(&wg, level)
	gpio := bbox.InitGpio(&wg, press)
	fish := leds.InitFish(&wg, level, press)

	go gpio.Run()
	go fish.Run()

	wg.Wait()
}
