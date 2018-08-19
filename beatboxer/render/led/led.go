package led

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
	log "github.com/sirupsen/logrus"
)

const (
	LED_FREQ     = 1200000
	LED_COUNT    = 300
	FREEFORM_IDX = 188
	TICK_DELAY   = 6 // match sound to LEDs
	SINE_PERIOD  = 5
)

type Led struct {
}

type Row struct {
	start   int
	end     int
	buttons [bbox.BEATS]int
}

var (
	rows = [bbox.SOUNDS]Row{
		// rows 0 and 1 are LED strip 0
		Row{
			start: 71,
			end:   0,
			buttons: [bbox.BEATS]int{
				68, 64, 60, 56,
				41, 37, 33, 29,
				27, 23, 19, 15,
				13, 9, 5, 1,
			},
		},
		Row{
			start: 72,
			end:   151,
			buttons: [bbox.BEATS]int{
				75, 79, 83, 88,
				103, 108, 112, 117,
				119, 124, 128, 133,
				136, 140, 145, 150,
			},
		},

		// rows 2 and 3 are LED strip 1
		Row{
			start: 83,
			end:   0,
			buttons: [bbox.BEATS]int{
				79, 74, 69, 64,
				53, 47, 42, 37,
				34, 29, 24, 18,
				16, 10, 5, 0,
			},
		},
		Row{
			start: 84,
			end:   176,
			buttons: [bbox.BEATS]int{
				88, 93, 99, 105,
				115, 121, 127, 133,
				136, 142, 148, 154,
				157, 163, 169, 174,
			},
		},
	}
)

func InitLed() *Led {
	log.Debugf("InitLed")
	leds.InitLeds(LED_FREQ, LED_COUNT, LED_COUNT)
	return &Led{}
}

func (w *Led) Render(state render.State) {
	// tick = (tick + l.iv.Ticks - TICK_DELAY) % l.iv.Ticks
	ws2811.Clear()

	for i := range state.Transitions {
		channel := 0
		if i > 1 {
			channel = 1
		}

		for j, tr := range state.Transitions[i] {
			if tr.Color != 0 {
				peak := rows[i].LocationToPeak(float64(j) + tr.Location)
				sineMap := color.GetSineVals(LED_COUNT, peak, SINE_PERIOD)

				log.Debugf("rows[%d].LocationToPeak(%f): %f", i, float64(j)+tr.Location, peak)
				log.Debugf("color.GetSineVals(%d, %f, %d): %f", LED_COUNT, peak, SINE_PERIOD, sineMap)

				for led, value := range sineMap {
					ws2811.SetLed(channel, led, color.MultiplyColor(tr.Color, float64(value)/255.0))
				}
			}
		}
	}

	actives := 0

	for i := range state.LEDs {
		channel := 0
		if i > 1 {
			channel = 1
		}

		for j, c := range state.LEDs[i] {
			if c != 0 {
				ws2811.SetLed(channel, rows[i].buttons[j], c)
				if c == color.ActiveBeatPurple {
					actives++
				}
			}
		}
	}

	// light freeform leds based on beat activity
	for i := 0; i < actives; i++ {
		r := rand.Intn(LED_COUNT - FREEFORM_IDX)
		ws2811.SetLed(0,
			r+FREEFORM_IDX,
			color.Colors[r%len(color.Colors)],
		)
		ws2811.SetLed(1,
			r+FREEFORM_IDX,
			color.Colors[(r+1)%len(color.Colors)],
		)
	}

	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}
}

// TODO: cache?
func (r *Row) LocationToPeak(location float64) float64 {
	f := math.Floor(location) // 12
	c := math.Ceil(location)  // 13

	var floor int
	var ceil int

	if f == 0 {
		// between start and first beat
		floor = r.start
		ceil = r.buttons[int(c)]
	} else if c == bbox.BEATS {
		// between last beat and end
		floor = r.buttons[int(f)]
		ceil = r.end
	} else {
		// between first and last beat
		floor = r.buttons[int(f)]
		ceil = r.buttons[int(c)]
	}

	percentAhead := location - f // 12.7 - 12 => 0.7
	diff := percentAhead * (float64(ceil) - float64(floor))

	return float64(floor) + diff
}
