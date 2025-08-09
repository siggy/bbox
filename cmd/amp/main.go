package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/program"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

func main() {
	logLevel := flag.String("log-level", "info", "set log level")
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

	// The number of bands is hardcoded in the equalizer.
	wavs, err := wavs.New(wavPath)
	if err != nil {
		log.Fatalf("wavs.New failed: %v", err)
	}
	defer wavs.Close()

	keyboard, err := keyboard.New(keyboard.KeyMapsPC)
	if err != nil {
		log.Fatalf("keyboard.New failed: %v", err)
	}
	presses := keyboard.Presses()
	go keyboard.Run()

	fmt.Println("Press 1,2,3,4 for beats, press 'r' for pyramid...")

	ticks := 0
	start := time.Now()
	last := time.Now()

	for {
		select {
		case press, ok := <-presses:
			if !ok {
				log.Info("keyboard channel closed, exiting...")
				return
			}
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

		case data, ok := <-wavs.EQ():
			ticks++

			t := time.Since(start)
			log.Infof("ticks: %d, time: %v, rate: %.2f/sec", ticks, t, float64(ticks)/t.Seconds())
			log.Infof("len(wavs.EQ()): %d", len(wavs.EQ()))

			log.Infof("time since last: %v", time.Since(last))
			last = time.Now()

			if !ok {
				continue
			}
			printDisplay(data)

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}

func printDisplay(data wavs.DisplayData) string {
	for i, band := range data {
		s := ""
		for range int(band * 100) {
			s += "█"
		}
		fmt.Printf("%d: %s\n", i, s)
	}

	return ""
}
