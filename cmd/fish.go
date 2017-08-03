package main

import (
	"sync"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/leds"
)

func main() {
	var wg sync.WaitGroup

	press := make(chan struct{})

	gpio := bbox.InitGpio(&wg, press)
	fish := leds.InitFish(&wg, press)

	go gpio.Run()
	go fish.Run()

	wg.Wait()
}
