package main

import (
	"log"
	"time"

	"github.com/siggy/bbox/bbox2"
)

func main() {
	wavs, err := bbox2.Init()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	// wavs.Play("perc-808.wav")
	wavs.Play("tom-808.wav")

	time.Sleep(2000 * time.Millisecond)
}
