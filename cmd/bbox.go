package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/leds"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// beat changes
	//   keyboard => loop
	//   keyboard => render
	//   keyboard => leds
	msgs := []chan bbox.Beats{
		make(chan bbox.Beats),
		make(chan bbox.Beats),
		make(chan bbox.Beats),
	}

	// tempo changes
	//	 keyboard => loop
	tempo := make(chan int)

	// ticks
	//   loop => render
	//   loop => leds
	ticks := []chan int{
		make(chan int),
		make(chan int),
	}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(bbox.WriteonlyBeats(msgs), tempo, false)
	loop := bbox.InitLoop(msgs[0], tempo, bbox.WriteonlyInt(ticks))
	render := bbox.InitRender(msgs[1], ticks[0])
	leds := leds.InitLedBeats(msgs[2], ticks[1])

	go keyboard.Run()
	go loop.Run()
	go render.Run()
	go leds.Run()

	defer keyboard.Close()
	defer loop.Close()
	defer render.Close()
	defer leds.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
