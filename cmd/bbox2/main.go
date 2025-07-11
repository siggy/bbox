package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
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
	stripLengths = []int{144}
)

func main() {
	logLevel := flag.String("log-level", "info", "set log level (debug, info, warn, error, fatal, panic)")
	bboxKB := flag.Bool("bbox-keyboard", false, "enable beatboxer keyboard")
	fakeLEDs := flag.Bool("fake-leds", false, "enable fake LEDs")
	macDevice := flag.Bool("mac-device", false, "connect to scorpio from a macbook (default is Raspberry Pi)")
	user := flag.String("user", "siggy-pi", "set user, eg siggy-pi, jhon-mac")
	eqBands := flag.Int("L", 16, "number of frequency bands (lights) for the EQ display")
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

	var wavPath string
	if *user != "jhon-mac" {
		wavPath = filepath.Join(os.Getenv("HOME"), "code", "bbox", "wavs")
	} else {
		wavPath = filepath.Join(os.Getenv("HOME"), "Documents", "bbox", "wavs")
	}

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
			if press.Col == program.Cols-1 && press.Row == program.Rows-1 {
				yieldCount++
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
			//old style led:
			log.Tracef("leds: %s", leds)
			ledStrips.Set(leds)

			select {
			case eqBandData := <-wavs.EQ():
				// show color bar:
				if len(eqBandData) == *eqBands {
					eqLine := buildEqLine(eqBandData, *eqBands)
					fmt.Println(eqLine)
				}
			default:
				// No new EQ data, do nothing.
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

// --- Display Functions ---

func magnitudeToColor(intensity float64) string {
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1 {
		intensity = 1
	}
	r := byte(intensity * 255)
	b := byte((1 - intensity) * 255)
	return fmt.Sprintf("\033[38;2;%d;0;%dm", r, b)
}


// show line of 16 blocks based on EQ data
// red is high intensity, blue is low intensity:
func buildEqLine(eqData []float64, eqBands int) string {
	var sb strings.Builder
	const resetColor = "\033[0m"

	sb.WriteString("EQ: [")
	for i := 0; i < eqBands; i++ {
		intensity := 0.0
		if i < len(eqData) {
			intensity = eqData[i]
		}
		sb.WriteString(magnitudeToColor(intensity))
		sb.WriteString("█")
		sb.WriteString(resetColor)
	}
	sb.WriteString("]")

	return sb.String()
}