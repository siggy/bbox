package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/programs"
	"github.com/siggy/bbox/bbox2/programs/beats"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

// keyboard -> beats -> ticks -> wavs
//                        -> leds
//
// keyboard
//   <-presses
// wavs
//   wavs.Play("filename.wav")
// buttons

func main() {
	logLevel := flag.String("log-level", "debug", "set log level (debug, info, warn, error, fatal, panic)")
	bboxKB := flag.Bool("bbox-keyboard", false, "enable beatboxer keyboard")
	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(lvl)

	keyMaps := keyboard.KeyMapsPC
	if *bboxKB {
		keyMaps = keyboard.KeyMapsRPI
	}

	// usb.Run()
	// os.Exit(0)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init
	wavs, err := wavs.New()
	if err != nil {
		log.Fatalf("wavs.New failed: %v", err)
	}
	defer wavs.Close()

	beats := beats.New()

	keyboard, err := keyboard.New(keyMaps)
	if err != nil {
		log.Fatalf("keyboard.New failed: %v", err)
	}

	presses := keyboard.Presses()

	// sound check

	for range 1 {
		wavs.Play("perc-808.wav")
		time.Sleep(100 * time.Millisecond)
		wavs.Play("hihat-808.wav")
		time.Sleep(100 * time.Millisecond)
		wavs.Play("kick-classic.wav")
		time.Sleep(100 * time.Millisecond)
		wavs.Play("tom-808.wav")
		time.Sleep(100 * time.Millisecond)
		wavs.Play("ceottk001_human.wav")
		time.Sleep(100 * time.Millisecond)
	}

	// run

	go keyboard.Run()
	go beats.Run()

	// TODO:
	// program interface {
	// 	Press() Coord<-
	//  Play() <-String
	//  Render() <-LEDs
	// }

	programs := []programs.Program{beats}
	cur := 0

	for {
		program := programs[cur]

		select {
		case press, more := <-presses:
			if !more {
				log.Info("keyboard channel closed, exiting...")
				return
			}
			log.Debugf("press: %q", press)
			program.Press(press)

		case leds := <-program.Render():
			log.Debugf("leds:\n%s", leds)

		case play := <-program.Play():
			log.Debugf("play: %s", play)
			wavs.Play(play)

		case <-program.Yield():
			log.Debugf("yield")
			cur = (cur + 1) % len(programs)

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}
