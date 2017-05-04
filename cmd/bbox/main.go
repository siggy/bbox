package main

import (
	"github.com/siggy/bbox/bbox"
)

func main() {
	msgs := make(chan bbox.Beats)

	keyboard := bbox.InitKeyboard(msgs)
	loop := bbox.InitLoop(msgs)

	// keyboard broadcasts quit with close(msgs)
	go keyboard.Run()

	// loop.Run() blocks until close(msgs)
	loop.Run()
}
