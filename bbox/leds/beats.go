package leds

import (
	"fmt"
	"sync"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	GPIO_PIN   = 18 // PWM0, must be 18 or 12
	LED_COUNT  = 30 // 144 * 5 // * 5 // * (1 + 5 + 5) // 30/m
	TICK_DELAY = 3  // match sound to LEDs
)

type Leds struct {
	beats bbox.Beats
	msgs  <-chan bbox.Beats
	ticks <-chan int
	wg    *sync.WaitGroup
}

func InitLeds(wg *sync.WaitGroup, msgs <-chan bbox.Beats, ticks <-chan int) *Leds {
	wg.Add(1)

	return &Leds{
		msgs:  msgs,
		ticks: ticks,
		wg:    wg,
	}
}

func (l *Leds) Run() {
	defer l.wg.Done()

	err := ws2811.Init(GPIO_PIN, LED_COUNT, BRIGHTNESS, 0, 0, 0)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	fmt.Printf("calling Clear()\n")
	ws2811.Clear()
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

	for {
		select {
		case tick := <-l.ticks:
			// TODO: leds for all 4 beats
			tick = (tick + bbox.BEATS - TICK_DELAY) % bbox.BEATS
			ws2811.Clear()
			ws2811.SetLed(0, tick, whitew)

			for _, beat := range l.beats {
				for j, t := range beat {
					if t {
						if j == tick {
							ws2811.SetLed(0, j, redw)
						} else {
							ws2811.SetLed(0, j, Red)
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
