package equalizer

import (
	"math"

	"github.com/siggy/bbox/pkg/leds"
)

// colorizer is a function type that maps a value from 0.0-1.0 to an LED color
type colorizer func(float64) leds.Color

const gamma = 6

var lutGammaMap = lutGamma(gamma)

func lutGamma(gamma float64) [256]uint8 {
	var lut [256]uint8
	inv := 1.0 / 255.0
	for x := range lut {
		y := 255 * math.Pow(float64(x)*inv, gamma)
		lut[x] = uint8(math.Round(y))
	}
	return lut
}

// Colorize constructs the new 4-bar spectrum history dashboard.
func Colorize(data DisplayData) [HistorySize][]leds.Color {
	// This exponent will be used to create a curve, making low values even lower.
	const exponent = 2.5

	colorizers := [HistorySize]colorizer{
		// 1. Electric Blue (Oldest)
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := lutGammaMap[uint8(0*scaledVal)], lutGammaMap[uint8(200*scaledVal)], lutGammaMap[uint8(255*scaledVal)]
			return leds.Color{R: r, G: g, B: b}
		},
		// 2. Toxic Green
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := lutGammaMap[uint8(100*scaledVal)], lutGammaMap[uint8(255*scaledVal)], lutGammaMap[uint8(100*scaledVal)]
			return leds.Color{R: r, G: g, B: b}
		},
		// 3. Sunset Orange
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := lutGammaMap[uint8(255*scaledVal)], lutGammaMap[uint8(150*scaledVal)], lutGammaMap[uint8(20*scaledVal)]
			return leds.Color{R: r, G: g, B: b}
		},
		// 4. Cyberpunk Pink (Newest)
		func(val float64) leds.Color {
			scaledVal := math.Pow(val, exponent)
			r, g, b := lutGammaMap[uint8(255*scaledVal)], lutGammaMap[uint8(50*scaledVal)], lutGammaMap[uint8(200*scaledVal)]
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
