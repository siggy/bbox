package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// beat changes
	//   keyboard => loop
	//   keyboard => render
	msgs := []chan bbox.Beats{
		make(chan bbox.Beats),
		make(chan bbox.Beats),
	}

	// ticks
	//   loop => render
	ticks := []chan int{
		make(chan int),
	}

	// keyboard broadcasts quit with close(msgs)
	keyboard := bbox.InitKeyboard(bbox.WriteonlyBeats(msgs), false)
	loop := bbox.InitLoop(msgs[0], bbox.WriteonlyInt(ticks))
	render := bbox.InitRender(msgs[1], ticks[0])

	go keyboard.Run()
	go loop.Run()
	go render.Run()

	defer keyboard.Close()
	defer loop.Close()
	defer render.Close()

	for {
		select {
		case <-sig:
			return
		}
	}
}
