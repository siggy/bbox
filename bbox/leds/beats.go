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
)

type Row struct {
	start   int
	end     int
	buttons [bbox.BEATS]int
}

// TODO: cache?
func (r *Row) TickToLed(tick int) (led int, buttonIdx int) {
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

	led = floor + int(diff)
	buttonIdx = -1
	if led == floor {
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
				31, 29, 27, 25, 23, 21, 19, 17, 15, 13, 11, 9, 7, 5, 3, 1,
			},
		},
		Row{
			start: 35,
			end:   69,
			buttons: [bbox.BEATS]int{
				36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66,
			},
		},

		// rows 2 and 3 are LED strip 1
		Row{
			start: 43,
			end:   0,
			buttons: [bbox.BEATS]int{
				42, 39, 36, 34, 31, 28, 26, 23, 20, 18, 15, 12, 10, 7, 4, 1,
			},
		},
		Row{
			start: 45,
			end:   87,
			buttons: [bbox.BEATS]int{
				46, 49, 52, 54, 57, 60, 63, 66, 68, 70, 73, 76, 78, 81, 84, 87,
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

			ledIdxs := [len(rows)]int{}
			buttonIdxs := [len(rows)]int{}
			for i, r := range rows {
				ledIdxs[i], buttonIdxs[i] = r.TickToLed(tick)
			}

			// light all leds at current position
			for i, _ := range rows[0:2] {
				ws2811.SetLed(0, ledIdxs[i], trueWhite)
			}
			for i, _ := range rows[2:4] {
				ws2811.SetLed(1, ledIdxs[i+2], trueWhite)
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
