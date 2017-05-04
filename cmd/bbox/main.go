package main

import (
	"io/ioutil"

	"github.com/siggy/bbox/bbox"
)

func main() {
	files, _ := ioutil.ReadDir("./wav")
	if len(files) != bbox.BEATS {
		panic(0)
	}

	msgs := make(chan bbox.Beats)

	keyboard := bbox.InitKeyboard(msgs)
	loop := bbox.InitLoop(msgs, files)

	// keyboard broadcasts quit with close(msgs)
	go keyboard.Run()

	// audio.Run() blocks until close(msgs)
	loop.Run()
}
