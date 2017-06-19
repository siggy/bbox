package main

import (
	"time"

	"github.com/siggy/bbox/bbox/leds"
)

func main() {
	leds.Init()
	for j := 0; j < 10000; j++ {
		for i := 0; i < leds.LED_COUNT; i++ {
			leds.SetLeds(i)
			time.Sleep(100 * time.Millisecond)
		}
	}
	leds.Shutdown()

	return
}
