package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/siggy/bbox/bbox2/equalizer"
	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

const ticksPerColorRotation = 15

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

	colorPos := 0
	colorTicks := 0

	ticks := 0
	start := time.Now()
	last := time.Now()

	for {
		select {
		case press, ok := <-presses:
			if !ok {
				log.Info("keyboard channel closed, exiting...")
				return
			}
			switch press {
			case program.Coord{Row: 0, Col: 0}:
				wavs.PlayWithEQ("hihat-808.wav")
			case program.Coord{Row: 0, Col: 1}:
				wavs.PlayWithEQ("kick-classic.wav")
			case program.Coord{Row: 0, Col: 2}:
				wavs.PlayWithEQ("perc-808.wav")
			case program.Coord{Row: 0, Col: 3}:
				wavs.PlayWithEQ("tom-808.wav")
			case program.Coord{Row: 1, Col: 3}:
				wavs.PlayWithEQ("pyramid.wav")
			}

		case data, ok := <-wavs.EQ():
			ticks++

			t := time.Since(start)
			log.Infof("ticks: %d, time: %v, rate: %.2f/sec", ticks, t, float64(ticks)/t.Seconds())
			log.Infof("len(wavs.EQ()): %d", len(wavs.EQ()))

			log.Infof("time since last: %v", time.Since(last))
			last = time.Now()

			if !ok {
				continue
			}
			fmt.Print(buildDisplay(data, colorPos))
			colorTicks++
			if colorTicks == ticksPerColorRotation {
				colorPos = (colorPos + 1) % equalizer.HistorySize // Cycle through colors
				colorTicks = 0
			}

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}

func colorizer(c leds.Color) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.R, c.G, c.B)
}

// Renders the 16-segment spectrum bar with a dynamic, per-segment color scheme.
func buildSpectrumBar(sb *strings.Builder, spectrum []leds.Color) {
	const resetColor = "\033[0m"
	const onBlock = "█"

	if len(spectrum) == 0 {
		// Draw an empty bar if there's no data
		sb.WriteString(strings.Repeat(" ", 16))
		sb.WriteString("\n")
		return
	}

	for _, norm := range spectrum {
		sb.WriteString(colorizer(norm))
		sb.WriteString(onBlock)
	}
	sb.WriteString(resetColor)
	sb.WriteString("\n")
}

// buildDisplay constructs the new 4-bar spectrum history dashboard.
func buildDisplay(data equalizer.DisplayData, colorPos int) string {
	colors := equalizer.Colorize(data)

	var sb strings.Builder

	sb.WriteString("\033[H") // Move cursor to home position

	// Render the 4 historical spectrum bars, from oldest to newest.
	for i := colorPos; i < colorPos+equalizer.HistorySize; i++ {
		buildSpectrumBar(&sb, colors[i%equalizer.HistorySize])
	}
	return sb.String()
}
