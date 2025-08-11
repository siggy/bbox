package leds

import (
	"fmt"
	"sort"
	"strings"
)

type (
	Color struct {
		R uint8 // 0-255
		G uint8 // 0-255
		B uint8 // 0-255
		W uint8 // 0-255, white channel for RGBW LEDs
	}

	// map[strip][pixel]Color
	State map[int]map[int]Color
)

var (
	Black = Color{R: 0, G: 0, B: 0, W: 0}
	Red   = Color{R: 255, G: 0, B: 0, W: 0}
	White = Color{R: 0, G: 0, B: 0, W: 255}

	pink       = Color{159, 0, 159, 93}
	trueBlue   = Color{0, 0, 255, 0}
	TrueBlue   = trueBlue
	red        = Color{210, 0, 50, 40}
	lightGreen = Color{0, 181, 115, 43}
	TrueRed    = Color{255, 0, 0, 0}
	TrueWhite  = Color{0, 0, 0, 255}
	purple     = Color{82, 0, 197, 52}
	Mint       = Color{R: 0, G: 170, B: 140, W: 0}
	trueGreen  = Color{0, 255, 0, 0}
	DeepPurple = Color{200, 0, 100, 0}
	Orange     = Color{255, 165, 0, 0}
	Yellow     = Color{255, 255, 0, 0}
	Cyan       = Color{0, 255, 255, 0}
	SkyBlue    = Color{135, 206, 235, 0}
	Gold       = Color{255, 215, 0, 80}

	// special color to tell beatboxer an active beat is occurring
	ActiveBeatPurple = Color{127, 127, 0, 127}
)

// TODO: cache?
func Brightness(c Color, brightness float64) Color {
	if brightness <= 0 {
		return Black
	} else if brightness >= 1 {
		return c
	}

	return Color{
		R: uint8(float64(c.R) * brightness),
		G: uint8(float64(c.G) * brightness),
		B: uint8(float64(c.B) * brightness),
		W: uint8(float64(c.W) * brightness),
	}
}

// TODO: cache?
func PulseColor(c Color, brightness float64) Color {
	// clamp to [0,1]
	if brightness <= 0 {
		return Black
	} else if brightness >= 1 {
		return White
	}

	if brightness <= 0.5 {
		// 0.0 → black, 0.5 → c at half brightness
		return Brightness(c, brightness)
	}

	// 0.5 → c at 0.5 brightness
	// 1.0 → pure white (W=255)
	t := (brightness - 0.5) / 0.5 // remap 0.5..1 → 0..1

	// fade RGBW of c (at half brightness) to W-only white
	start := Brightness(c, 0.5)
	end := Color{R: 0, G: 0, B: 0, W: 255}

	return Color{
		R: uint8(float64(start.R)*(1-t) + float64(end.R)*t),
		G: uint8(float64(start.G)*(1-t) + float64(end.G)*t),
		B: uint8(float64(start.B)*(1-t) + float64(end.B)*t),
		W: uint8(float64(start.W)*(1-t) + float64(end.W)*t),
	}
}

func (c Color) String() string {
	return fmt.Sprintf("%d,%d,%d,%d", c.R, c.G, c.B, c.W)
}

func (s State) ApplyState(incoming State) {
	for strip, stripLEDs := range incoming {
		for pixel, color := range stripLEDs {
			s.Set(strip, pixel, color)
		}
	}
}

func (s State) Set(strip int, pixel int, color Color) {
	if _, ok := s[strip]; !ok {
		s[strip] = make(map[int]Color)
	}
	s[strip][pixel] = color
}

func (s State) copy() State {
	copy := make(State)
	for strip, pixels := range s {
		copy[strip] = make(map[int]Color)
		for pixel, color := range pixels {
			copy[strip][pixel] = color
		}
	}

	return copy
}

func (s State) diff(newState State) State {
	diff := s.copy()

	for strip, pixels := range newState {
		if _, ok := diff[strip]; !ok {
			diff[strip] = make(map[int]Color)
		}
		for pixel, color := range pixels {
			if oldColor, ok := diff[strip][pixel]; !ok || oldColor != color {
				diff[strip][pixel] = color
			} else {
				// if the color is the same, we can remove it from the diff
				delete(diff[strip], pixel)
			}
		}
	}

	return diff
}

type flatState struct {
	strip int
	pixel int
	color Color
}

func (f flatState) String() string {
	return fmt.Sprintf("%dx%d:%s", f.strip, f.pixel, f.color)
}

func (s State) String() string {
	flat := []flatState{}
	for strip, pixels := range s {
		for pixel, color := range pixels {
			flat = append(flat, flatState{strip: strip, pixel: pixel, color: color})
		}
	}

	sort.Slice(flat, func(i, j int) bool {
		return flat[i].strip < flat[j].strip || (flat[i].strip == flat[j].strip && flat[i].pixel < flat[j].pixel)
	})

	strs := make([]string, len(flat))
	for i, f := range flat {
		strs[i] = f.String()
	}

	return strings.Join(strs, ",")
}
