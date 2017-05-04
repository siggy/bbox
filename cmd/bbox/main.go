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
	audio := bbox.InitAudio(msgs, files)

	quit := make(chan struct{})

	go keyboard.Run(quit)
	go audio.Run()

	<-quit
}
