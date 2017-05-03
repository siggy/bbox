package main

import (
	"io/ioutil"
	"time"

	"github.com/siggy/bbox/bbox"
)

func main() {
	files, _ := ioutil.ReadDir("./wav")

	bs := bbox.InitBeatState(len(files))

	// starter beat
	bs.Toggle(0, 0)
	bs.Toggle(0, 8)

	keyboard := bbox.InitKeyboard(bs)
	audio := bbox.InitAudio(bs, files)

	go keyboard.Run()
	go audio.Run()

	time.Sleep(60 * time.Second)
}
