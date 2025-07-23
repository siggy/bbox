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

	// State strip => pixel => Color
	State [][]Color
	// stateMap map[strip][pixel]Color
	stateMap map[int]map[int]Color
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
	deepPurple = Color{200, 0, 100, 0}

	// special color to tell beatboxer an active beat is occurring
	ActiveBeatPurple = Color{127, 127, 0, 127}
)

// TODO: cache?
func Brightness(c Color, brightness float64) Color {
	if brightness < 0 || brightness > 1 {
		return c
	}

	return Color{
		R: uint8(float64(c.R) * brightness),
		G: uint8(float64(c.G) * brightness),
		B: uint8(float64(c.B) * brightness),
		W: uint8(float64(c.W) * brightness),
	}
}

func (c Color) String() string {
	return fmt.Sprintf("%d,%d,%d,%d", c.R, c.G, c.B, c.W)
}

func NewState(stripLengths []int) State {
	state := make(State, len(stripLengths))
	for strip, length := range stripLengths {
		state[strip] = make([]Color, length)
		// for pixel := range state[strip] {
		// 	state[strip][pixel] = Color{0, 0, 0, 0} // initialize to black
		// }
	}
	return state
}

func (s State) ApplyState(incoming State) {
	for strip, stripLEDs := range incoming {
		for pixel, color := range stripLEDs {
			s.Set(strip, pixel, color)
		}
	}
}

func (s State) Set(strip int, pixel int, color Color) {
	s[strip][pixel] = color
}

func (s State) copy() State {
	s2 := make(State, len(s))
	for strip, pixels := range s {
		s2[strip] = make([]Color, len(pixels))
		copy(s2[strip], pixels)
	}

	return s2
}

// TODO: this should return a map for diff'ing?
func (s State) diff(newState State) stateMap {
	stateMap := make(stateMap)

	for strip, pixels := range newState {
		for pixel, color := range pixels {
			if s[strip][pixel] != color {
				if _, ok := stateMap[strip]; !ok {
					stateMap[strip] = make(map[int]Color)
				}
				stateMap[strip][pixel] = color
			}
		}
	}

	return stateMap
}

// TODO: this should return a map for diff'ing?
func (s State) ToMap() stateMap {
	stateMap := make(stateMap, len(s))
	for strip := range s {
		stateMap[strip] = make(map[int]Color, len(s[strip]))
	}

	for strip, pixels := range s {
		for pixel, color := range pixels {
			stateMap[strip][pixel] = color
		}
	}

	return stateMap
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
