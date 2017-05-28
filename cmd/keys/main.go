package main

import (
	"sync"

	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
)

func main() {
	defer termbox.Close()

	var wg sync.WaitGroup

	// beat changes
	//   keyboard => []
	msgs := []chan bbox.Beats{}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(&wg, writeonlyBeats(msgs), true)

	go keyboard.Run()

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
