package main

import (
	"os"
	"os/signal"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/programs/drums"
	"github.com/siggy/bbox/beatboxer/render/web"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	file, err := os.OpenFile("beatboxer_noleds.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	defer file.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	log.Debugf("InitHarness")
	harness := beatboxer.InitHarness(
		web.InitWeb(),
		bbox.KeyMapsPC,
	)
	log.Debugf("InitHarness complete")

	log.Debugf("Registering apps")
	harness.Register(&drums.DrumMachine{})
	// harness.Register(&ceottk.Ceottk{})
	log.Debugf("Registering apps complete")

	log.Debugf("Running harness")
	go harness.Run()

	for {
		select {
		case <-sig:
			log.Debugf("Received OS signal, exiting")
			return
		}
	}
}
