package main

import (
	"fmt"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// 2x side fins
	STRAND_COUNT1 = 5
	STRAND_LEN1   = 144

	// 1x top and back fins
	STRAND_COUNT2 = 10
	STRAND_LEN2   = 60

	LED_COUNT1 = STRAND_COUNT1 * STRAND_LEN1 // 5*144 // * 2x(5) // 144/m
	LED_COUNT2 = STRAND_COUNT2 * STRAND_LEN2 // 10*60 // * 1x(4 + 2 + 4) // 60/m
)

var (
	pink      = leds.MkColor(159, 0, 159, 93)
	trueBlue  = leds.MkColor(0, 0, 255, 0)
	red       = leds.MkColor(210, 0, 50, 40)
	green     = leds.MkColor(0, 181, 115, 43)
	trueRed   = leds.MkColor(255, 0, 0, 0)
	trueWhite = leds.MkColor(0, 0, 0, 255)
	purple    = leds.MkColor(82, 0, 197, 52)
	mint      = leds.MkColor(0, 27, 0, 228)
	trueGreen = leds.MkColor(0, 255, 0, 0)

	colors = []uint32{
		pink,
		trueBlue,
		red,
		green,
		trueRed,
		trueWhite,
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
			color := leds.MkColorWeight(color1, color2, weight)

			for j := 0; j < STRAND_LEN1; j++ {
				strand1[i*STRAND_LEN1+j] = color
			}
		}

		for i := 0; i < STRAND_COUNT2; i++ {
			color1 := colors[(iter+i)%len(colors)]
			color2 := colors[(iter+i+1)%len(colors)]
			color := leds.MkColorWeight(color1, color2, weight)

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
	leds.Init(LED_COUNT1, LED_COUNT2)
	run()
}
