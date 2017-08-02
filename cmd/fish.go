package main

import (
	"fmt"
	"math"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	GPIO_PIN1A = 18 // PWM0, must be 18 or 12
	GPIO_PIN1B = 12 // PWM0, must be 18 or 12
	GPIO_PIN2  = 13 // PWM1, must be 13 for rPI 3

	STRAND_COUNT1 = 5
	STRAND_LEN1   = 144
	STRAND_COUNT2 = 10
	STRAND_LEN2   = 60

	LED_COUNT1 = STRAND_COUNT1 * STRAND_LEN1 // 5*144 // * 2x(5) // 144/m
	LED_COUNT2 = STRAND_COUNT2 * STRAND_LEN2 // 10*60 // * 1x(4 + 2 + 4) // 60/m
)

// expects 0 <= [r,g,b,w] <= 255
func mkColor(r uint32, g uint32, b uint32, w uint32) uint32 {
	return uint32(b + g<<8 + r<<16 + w<<24)
}

func colorWeight(color1 uint32, color2 uint32, weight float64) uint32 {
	b1 := color1 & 0x000000ff
	g1 := (color1 & 0x0000ff00) >> 8
	r1 := (color1 & 0x00ff0000) >> 16
	w1 := (color1 & 0xff000000) >> 24

	b2 := color2 & 0x000000ff
	g2 := (color2 & 0x0000ff00) >> 8
	r2 := (color2 & 0x00ff0000) >> 16
	w2 := (color2 & 0xff000000) >> 24

	return mkColor(
		uint32(float64(r1)+float64(int32(r2)-int32(r1))*math.Min(1, weight*2)),
		uint32(float64(g1)+float64(int32(g2)-int32(g1))*math.Max(0, weight*2-1)),
		uint32(float64(b1)+float64(int32(b2)-int32(b1))*weight),
		uint32(float64(w1)+float64(int32(w2)-int32(w1))*weight),
	)
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

var (
	pink      = mkColor(159, 0, 159, 93)
	trueBlue  = mkColor(0, 0, 255, 0)
	red       = mkColor(210, 0, 50, 40)
	green     = mkColor(0, 181, 115, 43)
	trueRed   = mkColor(255, 0, 0, 0)
	purple    = mkColor(82, 0, 197, 52)
	mint      = mkColor(0, 27, 0, 228)
	trueGreen = mkColor(0, 255, 0, 0)

	colors = []uint32{
		pink,
		trueBlue,
		red,
		green,
		trueRed,
		purple,
		mint,
		trueGreen,
	}
)

func run() {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	iter := 0
	weight := float64(0)

	strand1 := make([]uint32, LED_COUNT1)
	strand2 := make([]uint32, LED_COUNT2)

	for {
		for i := 0; i < STRAND_COUNT1; i++ {
			color1 := colors[(iter+i)%len(colors)]
			color2 := colors[(iter+i+1)%len(colors)]
			color := colorWeight(color1, color2, weight)

			for j := 0; j < STRAND_LEN1; j++ {
				strand1[i*STRAND_LEN1+j] = color
			}
		}

		for i := 0; i < STRAND_COUNT2; i++ {
			color1 := colors[(iter+i)%len(colors)]
			color2 := colors[(iter+i+1)%len(colors)]
			color := colorWeight(color1, color2, weight)

			for j := 0; j < STRAND_LEN2; j++ {
				strand2[i*STRAND_LEN2+j] = color
			}
		}

		ws2811.SetBitmap(0, strand1)
		ws2811.SetBitmap(1, strand2)

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

		if weight < 1 {
			weight += 0.01
		} else {
			weight = 0
			iter = (iter + 1) % len(colors)
		}
	}
}

func main() {
	initLeds()
	run()
}
