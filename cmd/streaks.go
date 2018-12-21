package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT1 = 150
	LED_COUNT2 = 300

	LENGTH = 10
	SPEED  = 10.0
)

func main() {
	length := LENGTH
	speed := SPEED
	r := 200
	g := 0
	b := 100
	w := 0

	fmt.Printf("length: %+v\n", length)
	fmt.Printf("speed: %+v\n", speed)
	fmt.Printf("r: %+v\n", r)
	fmt.Printf("g: %+v\n", g)
	fmt.Printf("b: %+v\n", b)
	fmt.Printf("w: %+v\n", w)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	leds.InitLeds(leds.DEFAULT_FREQ, LED_COUNT1, LED_COUNT2)
	strand := make([]uint32, LED_COUNT1)

	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	streakLoc := 0.0

	black := color.Make(0, 0, 0, 0)

	for {
		select {
		case <-sig:
			return
		default:
			// streaks
			sineMap := color.GetSineVals(LED_COUNT1, streakLoc, int(length))
			for i, _ := range strand {
				strand[i] = black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand[led] = color.Make(
					uint32(multiplier*float64(r)),
					uint32(multiplier*float64(g)),
					uint32(multiplier*float64(b)),
					uint32(multiplier*float64(w)),
				)
			}

			streakLoc += speed
			if streakLoc >= LED_COUNT1 {
				streakLoc = 0
			}

			ws2811.SetBitmap(0, strand)
			ws2811.SetBitmap(1, strand)

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
}
