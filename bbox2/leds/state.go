package leds

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

func NewState() State {
	return State{}
}

func (s State) Set(strip int, pixel int, color Color) {
	if _, ok := s[strip]; !ok {
		s[strip] = make(map[int]Color)
	}
	s[strip][pixel] = color
}
