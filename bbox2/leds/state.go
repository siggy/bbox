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
	Mint       = Color{62, 180, 137, 0}
	trueGreen  = Color{0, 255, 0, 0}
	deepPurple = Color{200, 0, 100, 0}

	// special color to tell beatboxer an active beat is occurring
	ActiveBeatPurple = Color{127, 127, 0, 127}
)

func (c Color) String() string {
	return fmt.Sprintf("%d,%d,%d,%d", c.R, c.G, c.B, c.W)
}

func (s State) Set(strip int, pixel int, color Color) {
	if _, ok := s[strip]; !ok {
		s[strip] = make(map[int]Color)
	}
	s[strip][pixel] = color
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
