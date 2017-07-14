package main

import (
	"fmt"
	"math"
	"math/rand"
	// "time"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	GPIO_PIN1A = 18      // PWM0, must be 18 or 12
	GPIO_PIN1B = 12      // PWM0, must be 18 or 12
	GPIO_PIN2  = 13      // PWM1, must be 13 for rPI 3
	LED_COUNT1 = 144 * 5 // 144 * 5 // * 5 // * (1 + 5 + 5) // 30/m
	LED_COUNT2 = 60 * 10 // 60 * 10 // * 5 // * (4 + 2 + 4) // 60/m

	PI_FACTOR = math.Pi / (255. * 2.)
)

// expects 0 <= [r,g,b,w] <= 255
func mkColor(r uint32, g uint32, b uint32, w uint32) uint32 {
	return uint32(b + g<<8 + r<<16 + w<<24)
}

// maps midpoint 128 => 32 for brightness
func scale(x float64) uint32 {
	// y = 1000*(0.005333 * 4002473^(x/1000)-0.005333)
	return uint32(1000 * (0.005333*math.Pow(4002473., x/1000.) - 0.005333))
}

func initLeds() {

	// init once for each PIN1 (PWM0)
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(
		GPIO_PIN1A, LED_COUNT1, leds.BRIGHTNESS,
		GPIO_PIN2, LED_COUNT2, leds.BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	ws2811.Fini()

	err = ws2811.Init(
		GPIO_PIN1B, LED_COUNT1, leds.BRIGHTNESS,
		GPIO_PIN2, LED_COUNT2, leds.BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	// warm up
	for i := 0; i < LED_COUNT1; i += 30 {
		fmt.Printf("warmup GPIO1: %+v of %+v\n", i, LED_COUNT1)
		for j := 0; j < i; j++ {
			ws2811.SetLed(0, j, leds.Red)
		}

		err := ws2811.Render()
		if err != nil {
			fmt.Printf("ws2811.Render failed: %+v\n", err)
			panic(err)
		}
		err = ws2811.Wait()
		if err != nil {
			fmt.Printf("ws2811.Wait failed: %+v\n", err)
			panic(err)
		}
	}
	for i := 0; i < LED_COUNT2; i += 30 {
		fmt.Printf("warmup GPIO2: %+v of %+v\n", i, LED_COUNT2)
		for j := 0; j < i; j++ {
			ws2811.SetLed(1, j, leds.Red)
		}

		err := ws2811.Render()
		if err != nil {
			fmt.Printf("ws2811.Render failed: %+v\n", err)
			panic(err)
		}
		err = ws2811.Wait()
		if err != nil {
			fmt.Printf("ws2811.Wait failed: %+v\n", err)
			panic(err)
		}
	}
}

func run() {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	iter := 0

	// precompute random color rotation
	randColors := make([]uint32, LED_COUNT1)
	for i := 0; i < LED_COUNT1; i++ {
		randColors[i] = mkColor(0, uint32(rand.Int31n(256)), uint32(rand.Int31n(256)), uint32(rand.Int31n(256)))
	}

	for {
		for i := 0; i < LED_COUNT1; i++ {
			// blue := scale(255 * math.Sin(PI_FACTOR*float64(i+iter%256)))
			// fmt.Printf("BLUE: %+v\n", blue)
			// fmt.Printf("ITER: %+v\n", iter)
			// color := mkColor(0, uint32(rand.Int31n(256/8)), uint32(rand.Int31n(256)), uint32(rand.Int31n(256/8)))
			ws2811.SetLed(0, i, randColors[(i+iter)%LED_COUNT1])
		}

		for i := 0; i < LED_COUNT2; i++ {
			// color := leds.Colors[(iter+i)%len(leds.Colors)]
			// ws2811.SetLed(1, i, color)
			ws2811.SetLed(1, i, randColors[(i+iter)%LED_COUNT1])
		}

		err := ws2811.Render()
		if err != nil {
			fmt.Printf("ws2811.Render failed: %+v\n", err)
			panic(err)
		}

		err = ws2811.Wait()
		if err != nil {
			fmt.Printf("ws2811.Wait failed: %+v\n", err)
			panic(err)
		}

		// time.Sleep(1 * time.Millisecond)

		iter++
	}
}

func main() {
	initLeds()
	run()
}
