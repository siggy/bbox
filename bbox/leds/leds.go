package leds

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	BRIGHTNESS = 64 // 0-255
	LED_COUNT  = 30 // * (1 + 5 + 5) // 30/m
	GPIO_PIN   = 18
	TICK_DELAY = 3 // match sound to LEDs
)

var (
	red    = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x00})
	redw   = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x10})
	green  = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x00})
	greenw = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x10})
	blue   = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x00})
	bluew  = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x10})
	white  = binary.LittleEndian.Uint32([]byte{0x10, 0x10, 0x10, 0x00})
	whitew = binary.LittleEndian.Uint32([]byte{0x10, 0x10, 0x10, 0x10})

	colors = []uint32{red, redw, green, greenw, blue, bluew, white, whitew}
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

	err := ws2811.Init(GPIO_PIN, LED_COUNT, BRIGHTNESS)
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
			ws2811.SetLed(tick, whitew)

			for _, beat := range l.beats {
				for j, t := range beat {
					if t {
						if j == tick {
							ws2811.SetLed(j, redw)
						} else {
							ws2811.SetLed(j, red)
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

/*
 * Standalone functions to test all LEDs
 */
func Init() {
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(GPIO_PIN, LED_COUNT, BRIGHTNESS)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func SetLed(led int) {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.SetLed(%+v)\n", led)
	ws2811.SetLed(led, red)

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func SetLeds(led int) {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	for i, color := range colors {
		for j := 0; j < 5; j++ {
			index := (led + i + len(colors)*j) % LED_COUNT
			fmt.Printf("ws2811.SetLed(%+v, %+v)\n", index, color)
			ws2811.SetLed(index, color)
		}
	}

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func Shutdown() {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Fini()\n")
	ws2811.Fini()
}

// Turn off all LEDs
func Clear() {
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(GPIO_PIN, LED_COUNT, BRIGHTNESS)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Fini()\n")
	ws2811.Fini()
}
