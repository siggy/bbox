package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siggy/bbox/bbox2/beats"
	"github.com/siggy/bbox/bbox2/keyboard"
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
	bboxKB := flag.Bool("bbox-keyboard", false, "enable beatboxer keyboard")
	flag.Parse()

	keyMaps := keyboard.KeyMapsPC
	if *bboxKB {
		keyMaps = keyboard.KeyMapsRPI
	}

	log.SetLevel(log.DebugLevel)

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

	beatStates := beats.State()
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

	// var programs = []program{}

	for {
		select {
		case press, more := <-presses:
			if !more {
				log.Info("keyboard channel closed, exiting...")
				return
			}
			log.Debugf("press: %q", press)

			beats.Press(press)
		case beatState := <-beatStates:
			log.Debugf("beat state:\n%s", beatState)
		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}
