package main

import (
	"sync"

	"github.com/siggy/bbox/pkg/bbox"
	"github.com/siggy/bbox/pkg/leds"
)

func main() {
	var wg sync.WaitGroup

	// beat changes
	//   keyboard => loop
	//   keyboard => render
	//   keyboard => leds
	msgs := []chan bbox.Beats{
		make(chan bbox.Beats),
		make(chan bbox.Beats),
		make(chan bbox.Beats),
	}

	// ticks
	//   loop => render
	//   loop => leds
	ticks := []chan int{
		make(chan int),
		make(chan int),
	}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(&wg, bbox.WriteonlyBeats(msgs), false)
	loop := bbox.InitLoop(&wg, msgs[0], bbox.WriteonlyInt(ticks))
	render := bbox.InitRender(&wg, msgs[1], ticks[0])
	leds := leds.InitLeds(&wg, msgs[2], ticks[1])

	go keyboard.Run()
	go loop.Run()
	go render.Run()
	go leds.Run()

	wg.Wait()
}
