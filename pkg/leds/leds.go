package leds

import (
	"encoding/binary"
	"fmt"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT = 30
	GPIO_PIN  = 18
	LOOPS     = 10
)

var (
	red    = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x00})
	redw   = binary.LittleEndian.Uint32([]byte{0x10, 0x20, 0x00, 0x00})
	green  = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x00})
	greenw = binary.LittleEndian.Uint32([]byte{0x10, 0x00, 0x20, 0x00})
	blue   = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x00, 0x20})
	bluew  = binary.LittleEndian.Uint32([]byte{0x10, 0x00, 0x00, 0x20})
	white  = binary.LittleEndian.Uint32([]byte{0x00, 0x10, 0x10, 0x10})
	whitew = binary.LittleEndian.Uint32([]byte{0x10, 0x10, 0x10, 0x10})

	colors = []uint32{red, redw, green, greenw, blue, bluew, white, whitew}
)

type Leds struct {
	beats Beats
	msgs  <-chan Beats
	ticks <-chan int
}

func InitLeds(msgs <-chan Beats, ticks <-chan int) *Leds {
	return &Leds{
		msgs:  msgs,
		ticks: ticks,
	}
}

func (l *Leds) Run() {
	err := ws2811.Init(GPIO_PIN, LED_COUNT, 64)
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
			ws2811.SetLed(tick, redw)
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
				fmt.Printf("leds::l.msgs beats: %+v\n", beats)
				// r.beats = beats
				// r.Draw()
			} else {
				// closing
				fmt.Printf("leds::l.msgs closing %+v\n", beats)
				return
			}
		}
	}

	// color := 0
	// fmt.Printf("cycle LEDs\n")
	// for l := 0; l < LOOPS; l++ {
	// 	for i := 0; i < LED_COUNT; i++ {
	// 		ws2811.SetLed(i, colors[color%len(colors)])
	// 		color++

	// 		err = ws2811.Render()
	// 		if err != nil {
	// 			fmt.Printf("ws2811.Render failed: %+v\n", err)
	// 			panic(err)
	// 		}
	// 		err = ws2811.Wait()
	// 		if err != nil {
	// 			fmt.Printf("ws2811.Wait failed: %+v\n", err)
	// 			panic(err)
	// 		}
	// 	}
	// }

	// fmt.Printf("calling Clear()\n")
	// ws2811.Clear()

	// err = ws2811.Render()
	// if err != nil {
	// 	fmt.Printf("ws2811.Render failed: %+v\n", err)
	// 	panic(err)
	// }
	// err = ws2811.Wait()
	// if err != nil {
	// 	fmt.Printf("ws2811.Wait failed: %+v\n", err)
	// 	panic(err)
	// }
}

/*
 * Standalone function to turn off all LEDs
 */
func Clear() {
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(GPIO_PIN, LED_COUNT, 64)
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
