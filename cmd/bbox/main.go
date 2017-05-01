package main

import (
	"github.com/siggy/bbox/bbox"
	"time"
)

func main() {
	audio := bbox.Init()
	// go audio.Play(0)
	// go audio.Play(1)
	// go audio.Play(2)
	// go audio.Play(3)

	// go bbox.RunInput()
	// go bbox.RunAudio()

	i := 0
	for {
		time.Sleep(100 * time.Millisecond)
		audio.Play(i % 4)
		time.Sleep(1000 * time.Millisecond)
		audio.Play(i % 4)
		i++
	}
}
