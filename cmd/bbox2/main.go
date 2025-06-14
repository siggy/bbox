package main

import (
	"context"
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
	log.SetLevel(log.DebugLevel)

	// usb.Run()
	// os.Exit(0)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wavs, err := wavs.Init()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer wavs.Close()

	for range 1 {
		wavs.Play("perc-808.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("hihat-808.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("kick-classic.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("tom-808.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("ceottk001_human.wav")
		time.Sleep(250 * time.Millisecond)
	}

	keyboard, err := keyboard.New()
	if err != nil {
		log.Fatalf("keyboard.New failed: %v", err)
	}
	beats := beats.New(beats.KeyMapsPC)

	presses := keyboard.Presses()
	beatStates := beats.State()

	go keyboard.Run()
	go beats.Run()

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
