package main

import (
	"fmt"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// 2x undercarriage strands
	STRAND_COUNT1 = 4
	STRAND_LEN1   = 60

	LED_COUNT1 = STRAND_COUNT1 * STRAND_LEN1 // 4*60 // * 2x(4) // 60/m
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

	for {
		color1 := colors[(iter)%len(colors)]
		color2 := colors[(iter+1)%len(colors)]
		color := leds.MkColorWeight(color1, color2, weight)

		for i := 0; i < LED_COUNT1; i++ {
			strand1[i] = color
		}

		ws2811.SetBitmap(0, strand1)

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
			weight += 0.003
		} else {
			weight = 0
			iter = (iter + 1) % len(colors)
		}
	}
}

func main() {
	leds.InitLeds(LED_COUNT1, 0)
	run()
}
