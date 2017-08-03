package main

import (
	"github.com/siggy/bbox/bbox/leds"
)

const (
	LED_COUNT1 = 144 * 5 // 144 * 5 // * 5 // * (1 + 5 + 5) // 30/m
	LED_COUNT2 = 60 * 10 // 60 * 10 // * 5 // * (4 + 2 + 4) // 60/m
)

func main() {
	leds.Clear(LED_COUNT1, LED_COUNT2)
	return
}
