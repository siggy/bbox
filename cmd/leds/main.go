package main

import (
	"time"

	"github.com/siggy/bbox/pkg/leds"
)

func main() {
	leds.Init()
	for i := 0; i < leds.LED_COUNT; i++ {
		leds.SetLeds(i)
		time.Sleep(10 * time.Millisecond)
	}
	leds.Shutdown()

	return
}
