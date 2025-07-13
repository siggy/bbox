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

	"github.com/siggy/bbox/bbox2/keyboard"
	"github.com/siggy/bbox/bbox2/program"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

// this now uses a small denoiser neural net to get a speech only signal
// uses onnx runtime (python)
// todo: need a flag to enable/disable the denoiser in case we don't want to (or can't) do it

func main() {
	logLevel := flag.String("log-level", "info", "set log level")
	// barWidth := flag.Int("L", 16, "width of the metric display bars and number of spectrum bands")
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

	// The number of bands is now hardcoded in the equalizer, so we don't need to pass barWidth.
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

	// fmt.Println("Press keys to play sounds and view audio metrics...")
	// fmt.Print("\033[H\033[2J")

	fmt.Println("Press 1,2,3,4 for beats, press 'r' for pyramid...")

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

// Renders a single bar with a solid color, filling from the center outwards.
func buildBar(sb *strings.Builder, value float64, width int, color, label string) {
	const resetColor = "\033[0m"
	const offBlock = "░"
	const onBlock = "█"

	// Scale the 0.0-1.0 value to the number of lit blocks for one half of the bar.
	halfWidth := float64(width) / 2.0
	halfLights := int(value * halfWidth)

	// Determine the start and end bounds of the lit section.
	center := width / 2
	start := center - halfLights
	end := center + halfLights

	// For odd widths, the center block is shared, so adjust the end boundary.
	if width%2 != 0 {
		end--
	}

	// sb.WriteString(fmt.Sprintf("%-22s", label))
	sb.WriteString(color)
	for i := 0; i < width; i++ {
		// Light up the block if it's within the calculated range.
		if i >= start && i < end {
			sb.WriteString(onBlock)
		} else {
			sb.WriteString(offBlock)
		}
	}
	sb.WriteString(resetColor)
	sb.WriteString("\n")
}

// Renders the 16-segment spectrum bar with per-segment coloring.
func buildSpectrumBar(sb *strings.Builder, spectrum []float64, label string) {
	const resetColor = "\033[0m"
	const onBlock = "█"

	// sb.WriteString(fmt.Sprintf("%-22s", label))

	// The spectrum data is now pre-normalized from 0.0 to 1.0.
	for _, norm := range spectrum {
		// Map normalized value to a Black -> Red color spectrum.
		red := int(norm * 255)
		color := fmt.Sprintf("\033[38;2;%d;0;0m", red)

		sb.WriteString(color)
		sb.WriteString(onBlock)
	}
	sb.WriteString(resetColor)
	sb.WriteString("\n")
}

// buildDisplay constructs the new 4-bar metrics dashboard.
func buildDisplay(data wavs.DisplayData) string {
	var sb strings.Builder
	const colorKurtosis = "\033[38;2;255;255;0m"
	const colorLowEnergy = "\033[38;2;0;255;0m"
	const colorDenoised = "\033[38;2;255;0;255m"
	width := len(data.Spectrum)

	sb.WriteString("\033[H\033[2J") // Clear screen
	// sb.WriteString("--- Audio Signal Metrics ---\n\n")

	// Bar 1 (Top)
	buildBar(&sb, data.Kurtosis, width, colorKurtosis, "Peakiness (Kurtosis):")
	// Bar 2 (Second)
	buildSpectrumBar(&sb, data.Spectrum, "Log Spectrum (L->H):")
	// Bar 3 (Third)
	buildBar(&sb, data.LowFreqEnergy, width, colorLowEnergy, "Lows (<200Hz) Energy:")

	// sb.WriteString("\n--- Denoised Signal ---\n\n")
	// Bar 4 (Fourth)
	buildBar(&sb, data.DenoisedLevel, width, colorDenoised, "Speech Level:")

	return sb.String()
}