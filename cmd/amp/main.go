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
	// fmt.Print("\033[H\033[2J") // Clear screen on start

	ticks := 0
	start := time.Now()

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

			if !ok {
				continue
			}
			fmt.Print(buildDisplay(data))

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}

// colorizer is a function type that maps a value from 0.0-1.0 to an ANSI color string.
type colorizer func(float64) string

// Renders the 16-segment spectrum bar with a dynamic, per-segment color scheme.
func buildSpectrumBar(sb *strings.Builder, spectrum []float64, styler colorizer) {
	const resetColor = "\033[0m"
	const onBlock = "█"

	if len(spectrum) == 0 {
		// Draw an empty bar if there's no data
		sb.WriteString(strings.Repeat(" ", 16))
		sb.WriteString("\n")
		return
	}

	for _, norm := range spectrum {
		sb.WriteString(styler(norm))
		sb.WriteString(onBlock)
	}
	sb.WriteString(resetColor)
	sb.WriteString("\n")
}

// buildDisplay constructs the new 4-bar spectrum history dashboard.
func buildDisplay(data wavs.DisplayData) string {
	var sb strings.Builder

	// This exponent will be used to create a curve, making low values even lower.
	const exponent = 2.5

	colorizers := []colorizer{
		// 1. Electric Blue (Oldest)
		func(val float64) string {
			scaledVal := math.Pow(val, exponent)
			r, g, b := int(0*scaledVal), int(200*scaledVal), int(255*scaledVal)
			return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
		},
		// 2. Toxic Green
		func(val float64) string {
			scaledVal := math.Pow(val, exponent)
			r, g, b := int(100*scaledVal), int(255*scaledVal), int(100*scaledVal)
			return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
		},
		// 3. Sunset Orange
		func(val float64) string {
			scaledVal := math.Pow(val, exponent)
			r, g, b := int(255*scaledVal), int(150*scaledVal), int(20*scaledVal)
			return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
		},
		// 4. Cyberpunk Pink (Newest)
		func(val float64) string {
			scaledVal := math.Pow(val, exponent)
			r, g, b := int(255*scaledVal), int(50*scaledVal), int(200*scaledVal)
			return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
		},
	}

	sb.WriteString("\033[H") // Move cursor to home position

	// Render the 4 historical spectrum bars, from oldest to newest.
	for i := 0; i < 4; i++ {
		buildSpectrumBar(&sb, data.History[i], colorizers[i])
	}

	return sb.String()
}
