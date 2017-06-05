package main

import (
	"sync"

	"github.com/siggy/bbox/bbox"
)

func main() {
	var wg sync.WaitGroup

	// beat changes
	//   keyboard => loop
	//   keyboard => render
	msgs := []chan bbox.Beats{
		make(chan bbox.Beats),
		make(chan bbox.Beats),
	}

	// ticks
	//   loop => render
	ticks := []chan int{
		make(chan int),
	}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(&wg, bbox.WriteonlyBeats(msgs), false)
	loop := bbox.InitLoop(&wg, msgs[0], bbox.WriteonlyInt(ticks))
	render := bbox.InitRender(&wg, msgs[1], ticks[0])

	go keyboard.Run()
	go loop.Run()
	go render.Run()

	wg.Wait()
}
