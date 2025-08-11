package leds

import (
	"fmt"
	"sort"
	"strings"
)

type (
	// map[strip][pixel]Color
	State map[int]map[int]Color
)

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
