package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/pattern"
	"github.com/siggy/bbox/bbox/renderer/web"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// beat changes
	//   keyboard => loop
	//   keyboard => leds
	msgs := []chan bbox.Beats{
		make(chan bbox.Beats),
		make(chan bbox.Beats),
	}

	// tempo changes
	//	 keyboard => loop
	tempo := make(chan int)

	// ticks
	//   loop => leds
	ticks := []chan int{
		make(chan int),
	}

	// interval changes
	//   loop => leds
	intervals := []chan bbox.Interval{
		make(chan bbox.Interval),
	}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(bbox.WriteonlyBeats(msgs), tempo, bbox.KeyMapsPC, false)
	loop := bbox.InitLoop(msgs[0], tempo, bbox.WriteonlyInt(ticks), bbox.WriteonlyInterval(intervals))
	// leds := pattern.InitLedBeats(msgs[1], ticks[0], intervals[0], renderer.Screen{})
	leds := pattern.InitLedBeats(msgs[1], ticks[0], intervals[0], web.InitWeb())

	go keyboard.Run()
	go loop.Run()
	go leds.Run()

	// defer termbox.Close()
	defer keyboard.Close()
	defer loop.Close()
	defer leds.Close()

	for {
		select {
		case <-sig:
			return
		}
	}

	// termbox.Init() called in InitKeyboard()
}
