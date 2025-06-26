package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	"github.com/siggy/bbox/bbox2/programs/beats"
	"github.com/siggy/bbox/bbox2/programs/ledtest"
	"github.com/siggy/bbox/bbox2/programs/nice"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

const (
	yieldLimit = 5 // number of yields repetition keys before switching programs
)

var (
	// should match scorpio/code.py
	// StripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}
	stripLengths = []int{30}
)

func main() {
	logLevel := flag.String("log-level", "debug", "set log level (debug, info, warn, error, fatal, panic)")
	bboxKB := flag.Bool("bbox-keyboard", false, "enable beatboxer keyboard")
	fakeLEDs := flag.Bool("fake-leds", false, "enable fake LEDs")
	macDevice := flag.Bool("mac-device", false, "connect to scorpio from a macbook (default is Raspberry Pi)")
	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(lvl)
	log := log.WithField("bbox2", "main")

	keyMaps := keyboard.KeyMapsPC
	if *bboxKB {
		keyMaps = keyboard.KeyMapsRPI
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init
	wavs, err := wavs.New()
	if err != nil {
		log.Fatalf("wavs.New failed: %v", err)
	}
	defer wavs.Close()

	var ledStrips leds.LEDs
	if *fakeLEDs {
		ledStrips, err = leds.NewFake(stripLengths)
		if err != nil {
			log.Errorf("leds.NewFake failed: %+v", err)
			os.Exit(1)
		}
	} else {
		ledStrips, err = leds.New(stripLengths, *macDevice)
		if err != nil {
			log.Errorf("leds.New failed: %+v", err)
			os.Exit(1)
		}
	}
	defer ledStrips.Close()
	ledStrips.Clear()

	keyboard, err := keyboard.New(keyMaps)
	if err != nil {
		log.Fatalf("keyboard.New failed: %v", err)
	}
	presses := keyboard.Presses()

	go keyboard.Run()

	programs := []program.ProgramFactory{
		beats.NewProgram,
		ledtest.NewProgram,
		nice.NewProgram,
	}

	cur := 0
	progCtx, cancelProg := context.WithCancel(ctx)
	curProgram := programs[cur](progCtx)

	yieldCount := 0

	for {
		yield := func() {
			log.Debugf("yield prev program: %s", curProgram.Name())

			cancelProg()
			curProgram.Close()
			cur = (cur + 1) % len(programs)
			progCtx, cancelProg = context.WithCancel(ctx)
			curProgram = programs[cur](progCtx)
			ledStrips.Clear()
			wavs.StopAll()
			yieldCount = 0

			log.Debugf("yield new program: %s", curProgram.Name())
		}

		select {
		case press, ok := <-presses:
			if !ok {
				log.Info("keyboard channel closed, exiting...")

				cancelProg()
				curProgram.Close()
				return
			}

			log.Debugf("press: %q", press)

			if press.Col == program.Cols-1 && press.Row == program.Rows-1 {
				yieldCount++
				log.Debugf("yieldCount: %d", yieldCount)
				if yieldCount >= yieldLimit {
					yield()
					continue
				}
			} else {
				yieldCount = 0
			}

			curProgram.Press(press)

		case leds := <-curProgram.Render():
			log.Debugf("leds: %s", leds)

			err := ledStrips.Write(leds)
			if err != nil {
				log.Errorf("leds.Write failed: %v", err)
			}

		case play := <-curProgram.Play():
			log.Tracef("play: %s", play)

			wavs.Play(play)

		case <-curProgram.Yield():
			yield()

		case <-ctx.Done():
			log.Info("context done")

			cancelProg()
			curProgram.Close()
			return
		}
	}
}
