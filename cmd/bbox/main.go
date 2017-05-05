package main

import (
	"github.com/siggy/bbox/bbox"
)

func main() {
	// beat changes
	//   keyboard => loop
	//   keyboard => render
	msgs := []chan bbox.Beats{
		make(chan bbox.Beats),
		make(chan bbox.Beats),
	}

	// ticks
	//   loop => render
	ticks := make(chan int)

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(writeonly(msgs))
	loop := bbox.InitLoop(msgs[0], ticks)
	render := bbox.InitRender(msgs[1], ticks)

	go keyboard.Run()
	go render.Run()

	// loop.Run() blocks until close(msgs)
	loop.Run()
}

// can't pass a slice of non-direction channels as a slice of directional
// channels, so we have to convert the whole slice to directional first.
func writeonly(channels []chan bbox.Beats) []chan<- bbox.Beats {
	ret := make([]chan<- bbox.Beats, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}
