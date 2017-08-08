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
			tick = (tick + bbox.BEATS - TICK_DELAY) % bbox.BEATS
			ws2811.Clear()
			ws2811.SetLed(0, tick, trueWhite)

			for _, beat := range l.beats {
				for j, t := range beat {
					if t {
						if j == tick {
							ws2811.SetLed(0, j, redWhite)
						} else {
							ws2811.SetLed(0, j, trueRed)
						}
					}
				}
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
