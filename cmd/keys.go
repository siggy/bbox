package main

import (
	"sync"

	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/pkg/bbox"
)

func main() {
	defer termbox.Close()

	var wg sync.WaitGroup

	// beat changes
	//   keyboard => []
	msgs := []chan bbox.Beats{}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(&wg, bbox.WriteonlyBeats(msgs), true)

	go keyboard.Run()

	wg.Wait()
}
