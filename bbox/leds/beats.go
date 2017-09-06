package leds

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT    = 150
	FREEFORM_IDX = 100
	TICK_DELAY   = 17 // match sound to LEDs
	SINE_PERIOD  = 5
)

type Row struct {
	start   int
	end     int
	buttons [bbox.BEATS]int
}

// TODO: cache?
func (r *Row) TickToLed(tick int) (buttonIdx int, peak float64) {
	// determine where we are in the buttons array
	// 0 <= tick < 160
	// 0 <= beat < 16
	floatBeat := float64(tick) / float64(bbox.TICKS_PER_BEAT) // 127 => 12.7
	f := math.Floor(floatBeat)                                // 12
	c := math.Ceil(floatBeat)                                 // 13

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

	percentAhead := floatBeat - f // 12.7 - 12 => 0.7
	diff := percentAhead * (float64(ceil) - float64(floor))

	peak = float64(floor) + diff
	buttonIdx = -1
	if int(diff) == 0 {
		buttonIdx = int(f)
	}

	return
}

type LedBeats struct {
	beats   bbox.Beats
	closing chan struct{}
	msgs    <-chan bbox.Beats
	ticks   <-chan int
}

var (
	rows = [bbox.SOUNDS]Row{
		// rows 0 and 1 are LED strip 0
		Row{
			start: 33,
			end:   0,
			buttons: [bbox.BEATS]int{
				30, 28, 26, 24,
				22, 20, 18, 16,
				14, 12, 10, 8,
				6, 4, 2, 0,
			},
		},
		Row{
			start: 35,
			end:   69,
			buttons: [bbox.BEATS]int{
				33, 35, 38, 40,
				42, 45, 47, 49,
				52, 54, 57, 59,
				61, 63, 65, 68,
			},
		},

		// rows 2 and 3 are LED strip 1
		Row{
			start: 43,
			end:   0,
			buttons: [bbox.BEATS]int{
				38, 35, 33, 31,
				28, 25, 22, 20,
				18, 15, 13, 10,
				8, 5, 3, 0,
			},
		},
		Row{
			start: 45,
			end:   87,
			buttons: [bbox.BEATS]int{
				42, 45, 48, 51,
				54, 57, 60, 63,
				65, 68, 71, 74,
				77, 80, 83, 86,
			},
		},
	}
)

func InitLedBeats(msgs <-chan bbox.Beats, ticks <-chan int) *LedBeats {
	InitLeds(LED_COUNT, LED_COUNT)

	return &LedBeats{
		closing: make(chan struct{}),
		msgs:    msgs,
		ticks:   ticks,
	}
}

func (l *LedBeats) Run() {
	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	ws2811.Clear()
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

	for {
		select {
		case _, more := <-l.closing:
			if !more {
				fmt.Printf("LEDs closing\n")
				return
			}
		case tick := <-l.ticks:
			tick = (tick + bbox.TICKS - TICK_DELAY) % bbox.TICKS
			ws2811.Clear()

			buttonIdxs := [len(rows)]int{}
			peakVals := [len(rows)]float64{}
			for i, r := range rows {
				buttonIdxs[i], peakVals[i] = r.TickToLed(tick)
			}

			// light all leds around current position
			for i, _ := range rows[0:2] {
				sineMap := GetSineVals(LED_COUNT, peakVals[i], SINE_PERIOD)
				for led, value := range sineMap {
					ws2811.SetLed(0, led, MkColor(0, 0, 0, uint32(value)))
				}
			}
			for i, _ := range rows[2:4] {
				sineMap := GetSineVals(LED_COUNT, peakVals[i+2], SINE_PERIOD)
				for led, value := range sineMap {
					ws2811.SetLed(1, led, MkColor(0, 0, 0, uint32(value)))
				}
			}

			actives := 0

			// light active beats
			for i, beat := range l.beats[0:2] {
				for j, t := range beat {
					if t {
						if j == buttonIdxs[i] {
							actives++
							ws2811.SetLed(0, rows[i].buttons[j], purple)
						} else {
							ws2811.SetLed(0, rows[i].buttons[j], trueRed)
						}
					}
				}
			}
			for i, beat := range l.beats[2:4] {
				for j, t := range beat {
					if t {
						if j == buttonIdxs[i+2] {
							actives++
							ws2811.SetLed(1, rows[i+2].buttons[j], purple)
						} else {
							ws2811.SetLed(1, rows[i+2].buttons[j], trueRed)
						}
					}
				}
			}

			// light freeform leds based on beat activity
			for i := 0; i < actives; i++ {
				r := rand.Intn(LED_COUNT - FREEFORM_IDX)
				ws2811.SetLed(0,
					r+FREEFORM_IDX,
					Colors[r%len(Colors)],
				)
				ws2811.SetLed(1,
					r+FREEFORM_IDX,
					Colors[(r+1)%len(Colors)],
				)
			}

			err = ws2811.Render()
			if err != nil {
				fmt.Printf("ws2811.Render failed: %+v\n", err)
				panic(err)
			}
			err = ws2811.Wait()
			if err != nil {
				fmt.Printf("ws2811.Wait failed: %+v\n", err)
				panic(err)
			}
		case beats, more := <-l.msgs:
			if more {
				// incoming beat update from keyboard
				l.beats = beats
			} else {
				// closing
				fmt.Printf("LEDs closing\n")
				return
			}
		}
	}
}

func (l *LedBeats) Close() {
	// TODO: this doesn't block?
	close(l.closing)
}
