package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/program"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

type programScheduler struct {
	new    program.ProgramFactory
	code   []int
	hidden bool
}

const (
	yieldLimit = 5 // number of yields repetition keys before switching programs
)

var (
	// should match scorpio/code.py
	// StripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}
	stripLengths = []int{144}
)

func main() {
	logLevel := flag.String("log-level", "info", "set log level (debug, info, warn, error, fatal, panic)")
	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(lvl)
	log := log.WithField("bbox2", "main")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getwd failed: %v", err)
	}

	wavPath := filepath.Join(wd, "wavs")
	wavs, err := wavs.New(wavPath)
	if err != nil {
		log.Fatalf("wavs.New failed: %v", err)
	}
	defer wavs.Close()

	// wavs.Play("pyramid.wav")

	keyboard, err := keyboard.New(keyboard.KeyMapsPC)
	if err != nil {
		log.Fatalf("keyboard.New failed: %v", err)
	}
	presses := keyboard.Presses()

	go keyboard.Run()

	fmt.Println("Press 1,2,3,4 for beats, press 'r' for pyramid...")

	for {
		select {
		case press, ok := <-presses:
			if !ok {
				log.Info("keyboard channel closed, exiting...")
				return
			}
			log.Debugf("key pressed: %v", press)

			switch press {
			case program.Coord{Row: 0, Col: 0}:
				wavs.Play("hihat-808.wav")
			case program.Coord{Row: 0, Col: 1}:
				wavs.Play("kick-classic.wav")
			case program.Coord{Row: 0, Col: 2}:
				wavs.Play("perc-808.wav")
			case program.Coord{Row: 0, Col: 3}:
				wavs.Play("tom-808.wav")
			case program.Coord{Row: 1, Col: 3}:
				wavs.Play("pyramid.wav")
			}

		case eq, ok := <-wavs.EQ():
			if !ok {
				continue
			}

			log.Debugf("eq: %v", eq)

			for i, band := range eq {
				s := ""
				for range int(band * 10) {
					s += "█"
				}
				fmt.Printf("%d: %s\n", i, s)
			}

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}
