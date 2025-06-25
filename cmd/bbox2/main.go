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

var (
	// should match scorpio/code.py
	// StripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}
	stripLengths = []int{30}
)

func main() {
	logLevel := flag.String("log-level", "debug", "set log level (debug, info, warn, error, fatal, panic)")
	bboxKB := flag.Bool("bbox-keyboard", false, "enable beatboxer keyboard")
	fakeLEDs := flag.Bool("fake-leds", false, "enable fake LEDs")
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

	var ledStrips leds.LEDs
	if *fakeLEDs {
		ledStrips, err = leds.NewFake(stripLengths)
		if err != nil {
			log.Errorf("leds.NewFake failed: %+v", err)
			os.Exit(1)
		}
	} else {
		ledStrips, err = leds.New(stripLengths)
		if err != nil {
			log.Errorf("leds.New failed: %+v", err)
			os.Exit(1)
		}
	}
	defer ledStrips.Close()
	ledStrips.Clear()

	// beats := beats.New()

	keyboard, err := keyboard.New(keyMaps)
	if err != nil {
		log.Fatalf("keyboard.New failed: %v", err)
	}

	presses := keyboard.Presses()

	// sound check

	// for range 1 {
	// 	wavs.Play("perc-808.wav")
	// 	time.Sleep(100 * time.Millisecond)
	// 	wavs.Play("hihat-808.wav")
	// 	time.Sleep(100 * time.Millisecond)
	// 	wavs.Play("kick-classic.wav")
	// 	time.Sleep(100 * time.Millisecond)
	// 	wavs.Play("tom-808.wav")
	// 	time.Sleep(100 * time.Millisecond)
	// 	wavs.Play("ceottk001_human.wav")
	// 	time.Sleep(100 * time.Millisecond)
	// }

	// run

	go keyboard.Run()
	// go beats.Run()

	// TODO:
	// program interface {
	// 	Press() Coord<-
	//  Play() <-String
	//  Render() <-LEDs
	// }

	programs := []program.ProgramFactory{
		beats.NewProgram,
		ledtest.NewProgram,
		// other.NewProgram,
		// …
	}

	// programs := []programs.Program{beats}
	// index of the active program
	cur := 0
	// contexts and the running Program
	progCtx, cancelProg := context.WithCancel(ctx)
	program := programs[cur](progCtx)

	for {
		// program := programs[cur]

		select {
		case press, ok := <-presses:
			if !ok {
				log.Info("keyboard channel closed, exiting...")

				cancelProg()
				program.Close()
				return
			}

			log.Debugf("press: %q", press)

			program.Press(press)

		case leds := <-program.Render():
			log.Debugf("leds: %s", leds)

			err := ledStrips.Write(leds)
			if err != nil {
				log.Errorf("leds.Write failed: %v", err)
			}

		case play := <-program.Play():
			log.Tracef("play: %s", play)

			wavs.Play(play)

		case <-program.Yield():
			log.Debugf("yield: program %d", cur)

			// tear down old:
			cancelProg()
			program.Close()

			// pick next
			cur = (cur + 1) % len(programs)
			progCtx, cancelProg = context.WithCancel(ctx)
			program = programs[cur](progCtx)

		case <-ctx.Done():
			log.Info("context done")

			cancelProg()
			program.Close()
			return
		}
	}
}
