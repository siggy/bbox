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
	keyboard := bbox.InitKeyboard(&wg, writeonlyBeats(msgs), false)
	loop := bbox.InitLoop(&wg, msgs[0], writeonlyInt(ticks))
	render := bbox.InitRender(&wg, msgs[1], ticks[0])

	go keyboard.Run()
	go render.Run()
	go loop.Run()

	wg.Wait()
}

// can't pass a slice of non-direction channels as a slice of directional
// channels, so we have to convert the whole slice to directional first.
func writeonlyBeats(channels []chan bbox.Beats) []chan<- bbox.Beats {
	ret := make([]chan<- bbox.Beats, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}

func writeonlyInt(channels []chan int) []chan<- int {
	ret := make([]chan<- int, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}
