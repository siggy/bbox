package main

import (
	// _ "net/http/pprof"

	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/keyboard"
	"github.com/siggy/bbox/beatboxer/programs/ceottk"
	"github.com/siggy/bbox/beatboxer/programs/drums"
	"github.com/siggy/bbox/beatboxer/programs/intro"
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/render/led"
	"github.com/siggy/bbox/beatboxer/render/web"
)

func main() {
	bboxKB := flag.Bool("bbox-keyboard", false, "enable beatboxer keyboard")
	flag.Parse()

	// log.SetLevel(log.DebugLevel)
	// file, err := os.OpenFile("beatboxer_leds.log", os.O_CREATE|os.O_WRONLY, 0666)
	// if err == nil {
	// 	log.SetOutput(file)
	// } else {
	// 	log.Info("Failed to log to file, using default stderr")
	// }
	// defer file.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	keyMaps := bbox.KeyMapsPC
	if *bboxKB {
		keyMaps = bbox.KeyMapsRPI
	}

	harness := beatboxer.InitHarness(
		[]render.Renderer{
			web.InitWeb(),
			// render.InitTerminal(),
			led.InitLed(),
		},
		keyboard.Init(keyMaps),
	)

	harness.Register(&intro.Intro{})
	harness.Register(&drums.DrumMachine{})
	harness.Register(&ceottk.Ceottk{})

	go harness.Run()

	for {
		select {
		case <-sig:
			fmt.Printf("Received OS signal, exiting")
			return
		}
	}
}
