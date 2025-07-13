package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
	stripLengths = []int{144}
)

func main() {
	logLevel := flag.String("log-level", "info", "set log level (debug, info, warn, error, fatal, panic)")
	eqBands := flag.Int("L", 16, "number of frequency bands (lights) for the EQ display")

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

			// show log-scaled, dB-mapped color bar without DC bin:
			if len(eq) == *eqBands {
				eqLine := buildEqLine(eq, *eqBands)
				fmt.Println(eqLine)
			}

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}

// magnitudeToColor maps linear amplitude to a dB scale, then colors blue→red across a defined dB range
func magnitudeToColor(amplitude float64) string {
	if amplitude < 1e-6 {
		amplitude = 1e-6
	}
	dB := 20 * math.Log10(amplitude)
	const minDB = -60.0
	if dB < minDB {
		dB = minDB
	}
	ratio := (dB - minDB) / -minDB
	r := byte(ratio * 255)
	b := byte((1 - ratio) * 255)
	return fmt.Sprintf("\033[38;2;%d;0;%dm", r, b)
}

// buildEqLine constructs a log-frequency index color bar using magnitudes mapped via magnitudeToColor,
// skipping the DC bin (bin 0).
func buildEqLine(eqData []float64, eqBands int) string {
	var sb strings.Builder
	const resetColor = "\033[0m"
	m := len(eqData)

	sb.WriteString("EQ: [")
	for i := 0; i < eqBands; i++ {
		ratio := float64(i) / float64(eqBands-1)
		logIndex := math.Log10(1+9*ratio) / math.Log10(10)
		idx := int(logIndex*float64(m-1) + 0.5)
		// skip DC bin
		if idx < 1 {
			idx = 1
		} else if idx >= m {
			idx = m - 1
		}
		intensity := eqData[idx]
		sb.WriteString(magnitudeToColor(intensity))
		sb.WriteString("█")
		sb.WriteString(resetColor)
	}
	sb.WriteString("]")
	return sb.String()
}
