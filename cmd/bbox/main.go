package main

import (
	"github.com/siggy/bbox/bbox"
	"time"
)

func main() {
	go bbox.RunInput()
	go bbox.RunAudio()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
