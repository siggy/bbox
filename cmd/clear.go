package main

import (
	"github.com/siggy/bbox/bbox/leds"
)

const (
	GPIO_PIN1  = 18      // PWM0, must be 18 or 12
	GPIO_PIN2  = 13      // PWM1, must be 13 for rPI 3
	LED_COUNT1 = 144 * 5 // 144 * 5 // * 5 // * (1 + 5 + 5) // 30/m
	LED_COUNT2 = 60 * 10 // 60 * 10 // * 5 // * (4 + 2 + 4) // 60/m
)

func main() {
	leds.Clear(GPIO_PIN1, LED_COUNT1, GPIO_PIN2, LED_COUNT2)
	return
}
