package pattern

import (
	"fmt"
	"math"
	"math/rand"

	termbox "github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/renderer"
)

const (
	LED_FREQ     = 1200000
	LED_COUNT    = 300
	FREEFORM_IDX = 188
	TICK_DELAY   = 6 // match sound to LEDs
	SINE_PERIOD  = 5
)

type Row struct {
	start   int
	end     int
	buttons [bbox.BEATS]int
}

// TODO: cache?
func (r *Row) TickToLed(tick int, ticksPerBeat int) (buttonIdx int, peak float64) {
	// determine where we are in the buttons array
	// 0 <= tick < 160
	// 0 <= beat < 16
	floatBeat := float64(tick) / float64(ticksPerBeat) // 127 => 12.7
	f := math.Floor(floatBeat)                         // 12
	c := math.Ceil(floatBeat)                          // 13

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

	iv         bbox.Interval
	intervalCh <-chan bbox.Interval

	renderer.Renderer
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

func InitLedBeats(
	msgs <-chan bbox.Beats,
	ticks <-chan int,
	intervalCh <-chan bbox.Interval,
	r renderer.Renderer,
) *LedBeats {
	InitLeds(r, LED_FREQ, LED_COUNT, LED_COUNT)

	return &LedBeats{
		closing: make(chan struct{}),
		msgs:    msgs,
		ticks:   ticks,

		iv: bbox.Interval{
			TicksPerBeat: bbox.DEFAULT_TICKS_PER_BEAT,
			Ticks:        bbox.DEFAULT_TICKS,
		},
		intervalCh: intervalCh,

		Renderer: r,
	}
}

func (l *LedBeats) Run() {
	defer termbox.Close()
	defer func() {
		l.Clear()
		l.Render()
		l.Wait()
		l.Fini()
	}()

	l.Clear()
	err := l.Render()
	if err != nil {
		fmt.Printf("l.Render failed: %+v\n", err)
		panic(err)
	}
	err = l.Wait()
	if err != nil {
		fmt.Printf("l.Wait failed: %+v\n", err)
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
			tick = (tick + l.iv.Ticks - TICK_DELAY) % l.iv.Ticks
			l.Clear()

			buttonIdxs := [len(rows)]int{}
			peakVals := [len(rows)]float64{}
			for i, r := range rows {
				buttonIdxs[i], peakVals[i] = r.TickToLed(tick, l.iv.TicksPerBeat)
			}

			// light all leds around current position
			for i, _ := range rows[0:2] {
				sineMap := GetSineVals(LED_COUNT, peakVals[i], SINE_PERIOD)
				for led, value := range sineMap {
					l.SetLed(0, led, MkColor(0, 0, 0, uint32(value)))
				}
			}
			for i, _ := range rows[2:4] {
				sineMap := GetSineVals(LED_COUNT, peakVals[i+2], SINE_PERIOD)
				for led, value := range sineMap {
					l.SetLed(1, led, MkColor(0, 0, 0, uint32(value)))
				}
			}

			actives := 0

			// light active beats
			for i, beat := range l.beats[0:2] {
				for j, t := range beat {
					if t {
						if j == buttonIdxs[i] {
							actives++
							l.SetLed(0, rows[i].buttons[j], purple)
						} else {
							l.SetLed(0, rows[i].buttons[j], trueRed)
						}
					}
				}
			}
			for i, beat := range l.beats[2:4] {
				for j, t := range beat {
					if t {
						if j == buttonIdxs[i+2] {
							actives++
							l.SetLed(1, rows[i+2].buttons[j], purple)
						} else {
							l.SetLed(1, rows[i+2].buttons[j], trueRed)
						}
					}
				}
			}

			// light freeform leds based on beat activity
			for i := 0; i < actives; i++ {
				r := rand.Intn(LED_COUNT - FREEFORM_IDX)
				l.SetLed(0,
					r+FREEFORM_IDX,
					Colors[r%len(Colors)],
				)
				l.SetLed(1,
					r+FREEFORM_IDX,
					Colors[(r+1)%len(Colors)],
				)
			}

			err = l.Render()
			if err != nil {
				fmt.Printf("l.Render failed: %+v\n", err)
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
		case iv, more := <-l.intervalCh:
			if more {
				// incoming interval update from loop
				l.iv = iv
			} else {
				// we should never get here
				fmt.Printf("unexpected: intervalCh return no more\n")
				return
			}
		}
	}
}

func (l *LedBeats) Close() {
	// TODO: this doesn't block?
	close(l.closing)
}
