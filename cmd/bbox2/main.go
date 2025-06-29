package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	"github.com/siggy/bbox/bbox2/programs/beats"
	"github.com/siggy/bbox/bbox2/programs/devil"
	"github.com/siggy/bbox/bbox2/programs/ledtest"
	"github.com/siggy/bbox/bbox2/programs/nice"
	"github.com/siggy/bbox/bbox2/programs/pyramid"
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
	wavs, err := wavs.New("/home/sig/code/bbox/wavs")
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

	programs := []programScheduler{
		{new: beats.New, code: nil, hidden: false},
		{new: ledtest.New, code: nil, hidden: false},
		{new: nice.New, code: []int{1, 2, 1, 0}, hidden: true},
		{new: devil.New, code: []int{0, 9, 1, 7}, hidden: true},
		{new: pyramid.New, code: []int{0, 6, 0, 4}, hidden: true},
	}

	cur := 0
	progCtx, cancelProg := context.WithCancel(ctx)
	curProgram := programs[cur].new(progCtx)

	yieldCount := 0

	rollingCode := []int{0, 0, 0, 0}

	for {
		yield := func(next program.ProgramFactory) {
			log.Debugf("yield prev program: %s", curProgram.Name())

			cancelProg()
			curProgram.Close()

			if next == nil {
				for {
					cur = (cur + 1) % len(programs)
					if !programs[cur].hidden {
						break
					}
				}

				next = programs[cur].new
			} else {
				// back next up one so we yield back to the same place
				cur = (cur - 1 + len(programs)) % len(programs)
			}

			progCtx, cancelProg = context.WithCancel(ctx)
			curProgram = next(progCtx)

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

			// TODO: combine these two code patterns?
			if press.Col == program.Cols-1 && press.Row == program.Rows-1 {
				yieldCount++
				log.Debugf("yieldCount: %d", yieldCount)
				if yieldCount >= yieldLimit {
					yield(nil)
					continue
				}
			} else {
				yieldCount = 0
			}

			if press.Row == 0 {
				rollingCode = append(rollingCode[1:], press.Col)

				var found program.ProgramFactory

				for _, p := range programs {
					if slices.Equal(p.code, rollingCode) {
						log.Debugf("rolling code matched: %v", rollingCode)
						found = p.new
						break
					}
				}

				if found != nil {
					yield(found)
				}
			} else {
				rollingCode = []int{0, 0, 0, 0}
			}

			curProgram.Press(press)

		case leds := <-curProgram.Render():
			log.Tracef("leds: %s", leds)

			err := ledStrips.Write(leds)
			if err != nil {
				log.Errorf("leds.Write failed: %v", err)
			}

		case play := <-curProgram.Play():
			log.Tracef("play: %s", play)

			wavs.Play(play)

		case <-curProgram.Yield():
			yield(nil)

		case <-ctx.Done():
			log.Info("context done")

			cancelProg()
			curProgram.Close()
			return
		}
	}
}
