package leds

import (
	"fmt"
	"math"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT  = 150
	TICK_DELAY = 0 // match sound to LEDs
)

type Row struct {
	start   int
	end     int
	buttons [bbox.BEATS]int
}

func (r *Row) TickToLed(tick int) int {
	// determine where we are in the buttons array
	// 0 <= tick < 160
	// 0 <= beat < 16
	floatBeat := float64(tick) / float64(bbox.TICKS_PER_BEAT) // 12.7 => 0.7
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

	percentAhead := floatBeat - f
	diff := percentAhead * (float64(ceil) - float64(floor))
	return floor + int(diff)
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

		// rows 1 and 2 are LED strip 1
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
			// TODO: leds for all 4 beats
			tick = (tick + bbox.TICKS - TICK_DELAY) % bbox.TICKS
			ws2811.Clear()
			// cur := tick / bbox.TICKS_PER_BEAT

			// light all leds at current position
			for _, r := range rows[0:2] {
				ws2811.SetLed(0, r.TickToLed(tick), trueWhite)
			}
			for _, r := range rows[2:4] {
				ws2811.SetLed(1, r.TickToLed(tick), trueWhite)
			}

			// light active beats
			// for _, beat := range l.beats {
			// 	for j, t := range beat {
			// 		if t {
			// 			for _, r := range rows {
			// 				if j == cur {
			// 					ws2811.SetLed(0, r.buttons[j], redWhite)
			// 				} else {
			// 					ws2811.SetLed(0, r.buttons[j], trueRed)
			// 				}
			// 			}
			// 		}
			// 	}
			// }

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
