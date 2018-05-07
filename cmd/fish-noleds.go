package main

import (
	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/pattern"
	"github.com/siggy/bbox/bbox/renderer"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.SetOutputMode(termbox.Output256)

	level := make(chan float64)
	press := make(chan struct{})

	amplitude := bbox.InitAmplitude(level)
	fish := pattern.InitFish(renderer.Screen{}, level, press)

	go amplitude.Run()
	go fish.Run()

	defer amplitude.Close()
	defer fish.Close()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeySpace:
				go func() { press <- struct{}{} }()
			case termbox.KeyEsc:
				break loop
			}
		}
	}
}
