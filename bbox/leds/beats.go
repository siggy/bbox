package leds

import (
	"fmt"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT  = 150
	TICK_DELAY = 3 // match sound to LEDs
)

var (
	rows = [bbox.SOUNDS]Row{
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

type Row struct {
	start   int
	end     int
	buttons [bbox.BEATS]int
}

type LedBeats struct {
	beats   bbox.Beats
	closing chan struct{}
	msgs    <-chan bbox.Beats
	ticks   <-chan int
}

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
			beat := tick / bbox.TICKS_PER_BEAT

			for _, r := range rows {
				ws2811.SetLed(0, r.buttons[beat], trueWhite)
				ws2811.SetLed(1, r.buttons[beat], trueWhite)
			}

			// for _, beat := range l.beats {
			// 	for j, t := range beat {
			// 		if t {
			// 			for _, r := range rows {
			// 				if j == tick {
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
