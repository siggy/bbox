// usage:
//   ./streaks length speed r g b w

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT = 150
)

func parseArg(arg string) uint32 {
	tmp, _ := strconv.ParseUint(arg, 10, 32)
	return uint32(tmp)
}

func main() {
	args := os.Args[1:]
	length := parseArg(args[0])
	speed, _ := strconv.ParseFloat(args[1], 10)
	r := parseArg(args[2])
	g := parseArg(args[3])
	b := parseArg(args[4])
	w := parseArg(args[5])

	fmt.Printf("length: %+v\n", length)
	fmt.Printf("speed: %+v\n", speed)
	fmt.Printf("r: %+v\n", r)
	fmt.Printf("g: %+v\n", g)
	fmt.Printf("b: %+v\n", b)
	fmt.Printf("w: %+v\n", w)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	leds.InitLeds(LED_COUNT, LED_COUNT)
	strand := make([]uint32, LED_COUNT)

	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	streakLoc := 0.0

	black := leds.MkColor(0, 0, 0, 0)

	for {
		select {
		case <-sig:
			return
		default:
			// streaks
			sineMap := leds.GetSineVals(LED_COUNT, streakLoc, int(length))
			for i, _ := range strand {
				strand[i] = black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand[led] = leds.MkColor(
					uint32(multiplier*float64(r)),
					uint32(multiplier*float64(g)),
					uint32(multiplier*float64(b)),
					uint32(multiplier*float64(w)),
				)
			}

			streakLoc += speed
			if streakLoc >= LED_COUNT {
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
