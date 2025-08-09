package equalizer

import (
	"math"

	"github.com/siggy/bbox/bbox2/leds"
)

// colorizer is a function type that maps a value from 0.0-1.0 to an LED color
type colorizer func(float64) leds.Color

// Colorize constructs the new 4-bar spectrum history dashboard.
func Colorize(data DisplayData) [HistorySize][]leds.Color {
	// This exponent will be used to create a curve, making low values even lower.
	const exponent = 2.5

	colorizers := [HistorySize]colorizer{
		// 1. Electric Blue (Oldest)
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := uint8(0*scaledVal), uint8(200*scaledVal), uint8(255*scaledVal)
			return leds.Color{R: r, G: g, B: b}
		},
		// 2. Toxic Green
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := uint8(100*scaledVal), uint8(255*scaledVal), uint8(100*scaledVal)
			return leds.Color{R: r, G: g, B: b}
		},
		// 3. Sunset Orange
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := uint8(255*scaledVal), uint8(150*scaledVal), uint8(20*scaledVal)
			return leds.Color{R: r, G: g, B: b}
		},
		// 4. Cyberpunk Pink (Newest)
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := uint8(255*scaledVal), uint8(50*scaledVal), uint8(200*scaledVal)
			return leds.Color{R: r, G: g, B: b}
		},
	}

	colors := [HistorySize][]leds.Color{}

	// Render the 4 historical spectrum bars, from oldest to newest.
	for i := range HistorySize {
		colors[i] = make([]leds.Color, 16) // 16 LEDs per bar
		for j := 0; j < 16; j++ {
			norm := data[i][j]
			colors[i][j] = colorizers[i](norm)
		}
	}
	return colors
}
