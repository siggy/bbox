package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/siggy/bbox/pkg/keyboard"
	"github.com/siggy/bbox/pkg/leds"
	"github.com/siggy/bbox/pkg/program"
	"github.com/siggy/bbox/pkg/programs/beats"
	"github.com/siggy/bbox/pkg/programs/song"
	"github.com/siggy/bbox/pkg/wavs"
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
	stripLengths = []int{144, 144, 144, 144, 144, 144, 144, 144}

	defaultSounds = [program.Rows]string{
		"hihat-808.wav",
		"kick-classic.wav",
		"perc-808.wav",
		"tom-808.wav",
	}

	// if 75% of beats are active, yield to the next program
	defaultBeatLimit  = program.Rows * program.Cols * 3 / 4
	oneThirdBeatLimit = program.Rows * program.Cols / 3
)

func main() {
	logLevel := flag.String("log-level", "debug", "set log level (debug, info, warn, error, fatal, panic)")
	bboxKB := flag.Bool("bbox-keyboard", true, "enable beatboxer keyboard")
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
	wavPath := filepath.Join(os.Getenv("HOME"), "code", "bbox", "wavs")
	wavs, err := wavs.New(wavPath)
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
		ledStrips, err = leds.New(ctx, stripLengths, *macDevice)
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

	// TODO: first program should not be marked hidden
	programs := []programScheduler{
		{
			new: beats.New(
				"default",
				leds.Red, leds.White,
				defaultSounds,
				[]program.Coord{
					{Row: 1, Col: 0},
					{Row: 1, Col: 8},
				},
				120,
				defaultBeatLimit,
			),
			code: nil, hidden: false,
		},

		// We Will Rock You — Queen (thick: add kick under stomps)
		{
			new: beats.New(
				"we will rock you",
				leds.White, leds.Gold,
				[program.Rows]string{
					"clap-analog.wav",    // bright accent layer for clap
					"kick-stomp.wav",     // stomp core
					"clap-fat.wav",       // main clap
					"tom-acoustic01.wav", // low tom for stomp layer
				},
				[]program.Coord{
					// Stomps = kick + tom stacked
					{Row: 1, Col: 0}, {Row: 3, Col: 0},
					{Row: 1, Col: 4}, {Row: 3, Col: 4},

					// Flam into the clap (slightly early) + bright layer
					{Row: 2, Col: 9}, // main clap
					{Row: 0, Col: 9}, // bright layer
				},
				165,
				oneThirdBeatLimit,
			),
			code: nil, hidden: false,
		},

		// Stayin’ Alive — Bee Gees (disco: four-on-the-floor + 8th hats)
		{
			new: beats.New(
				"stayin alive",
				leds.Red, leds.Mint,
				[program.Rows]string{
					"hihat-acoustic02.wav", // Row 0 - bright disco hihat
					"kick-classic.wav",     // Row 1 - disco kick
					"snare-big.wav",        // Row 2 - roomy snare
					"openhat-tight.wav",    // Row 3 - open hihat accent
				},
				[]program.Coord{
					// Hihat (8ths)
					{Row: 0, Col: 0}, {Row: 0, Col: 2}, {Row: 0, Col: 4}, {Row: 0, Col: 6},
					{Row: 0, Col: 8}, {Row: 0, Col: 10}, {Row: 0, Col: 12}, {Row: 0, Col: 14},
					// Kick on all quarters
					{Row: 1, Col: 0}, {Row: 1, Col: 4}, {Row: 1, Col: 8}, {Row: 1, Col: 12},
					// Snare on 2 & 4
					{Row: 2, Col: 4}, {Row: 2, Col: 12},
				},
				104,
				oneThirdBeatLimit,
			),
			code: nil, hidden: false,
		},

		// Shape of You — Ed Sheeran (tight pop groove)
		{
			new: beats.New(
				"shape of you",
				leds.Orange, leds.Cyan,
				[program.Rows]string{
					"hihat-acoustic02.wav", // Row 0 - clean hihat
					"kick-classic.wav",     // Row 1 - main kick
					"snare-acoustic02.wav", // Row 2 - snare/clap blend
					"perc-tambo.wav",       // Row 3 - tambourine accent
				},
				[]program.Coord{
					// Hihat (8ths)
					{Row: 0, Col: 0}, {Row: 0, Col: 2}, {Row: 0, Col: 4}, {Row: 0, Col: 6},
					{Row: 0, Col: 8}, {Row: 0, Col: 10}, {Row: 0, Col: 12}, {Row: 0, Col: 14},
					// Kick
					{Row: 1, Col: 0}, {Row: 1, Col: 8}, {Row: 1, Col: 11},
					// Snare
					{Row: 2, Col: 4}, {Row: 2, Col: 12},
				},
				96,
				oneThirdBeatLimit,
			),
			code: nil, hidden: false,
		},

		// Four-on-the-floor (house)
		{
			new: beats.New(
				"four on the floor",
				leds.Red, leds.White,
				[program.Rows]string{
					"hihat-analog.wav",  // Row 0 - tight closed hihat
					"kick-classic.wav",  // Row 1 - punchy house kick
					"clap-808.wav",      // Row 2 - snappy clap/snare
					"openhat-tight.wav", // Row 3 - short open hihat
				},
				[]program.Coord{
					// Hihat (8ths)
					{Row: 0, Col: 0}, {Row: 0, Col: 2}, {Row: 0, Col: 4}, {Row: 0, Col: 6},
					{Row: 0, Col: 8}, {Row: 0, Col: 10}, {Row: 0, Col: 12}, {Row: 0, Col: 14},
					// Kick
					{Row: 1, Col: 0}, {Row: 1, Col: 8},
					// Snare-ish (perc)
					{Row: 2, Col: 4}, {Row: 2, Col: 12},
				},
				124,
				oneThirdBeatLimit,
			),
			code: nil, hidden: false,
		},

		// Dembow / reggaeton
		{
			new: beats.New(
				"dembow",
				leds.Yellow, leds.Cyan,
				[program.Rows]string{
					"hihat-acoustic01.wav", // Row 0 - bright hat
					"kick-classic.wav",     // Row 1 - punchy kick
					"snare-analog.wav",     // Row 2 - rimshot/snare tone
					"perc-tambo.wav",       // Row 3 - tambourine accent
				},
				[]program.Coord{
					// Hihat (8ths)
					{Row: 0, Col: 0}, {Row: 0, Col: 2}, {Row: 0, Col: 4}, {Row: 0, Col: 6},
					{Row: 0, Col: 8}, {Row: 0, Col: 10}, {Row: 0, Col: 12}, {Row: 0, Col: 14},
					// Kick
					{Row: 1, Col: 0}, {Row: 1, Col: 8},
					// Snare-ish (perc)
					{Row: 2, Col: 4}, {Row: 2, Col: 10},
				},
				100,
				oneThirdBeatLimit,
			),
			code: nil, hidden: false,
		},

		{new: song.New("wouldnt_it_be_nice.wav", time.Second*154), code: []int{1, 2, 1, 0}, hidden: true},
		{new: song.New("runnin_with_the_devil.wav", time.Second*215), code: []int{0, 9, 1, 7}, hidden: true},
		{new: song.New("pyramid.wav", time.Second*289), code: []int{0, 6, 0, 4}, hidden: true},
		{new: song.New("black_sabbath.wav", time.Second*378), code: []int{0, 4, 2, 0}, hidden: true},
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

		case displayData, ok := <-wavs.EQ():
			if !ok {
				log.Info("wavs.EQ() channel closed")
				continue
			}

			curProgram.EQ(displayData)

		case leds := <-curProgram.Render():
			log.Tracef("leds: %s", leds)

			ledStrips.Set(leds)

		case play := <-curProgram.Play():
			log.Tracef("play: %s", play)

			wavs.Play(play)

		case play := <-curProgram.PlayWithEQ():
			log.Tracef("play with eq: %s", play)

			wavs.PlayWithEQ(play)

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
