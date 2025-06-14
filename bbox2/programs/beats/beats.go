package beats

import (
	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/programs"
)

const (
	sounds = 4
	beats  = 16
)

type (
	BeatState [sounds][beats]bool

	Beats struct {
		beats    BeatState
		coordsCh chan programs.Coord
		// beatsCh  chan BeatState
		playCh chan string
		ledsCh chan leds.LEDs
	}
)

func (b BeatState) String() string {
	var str string
	for row := range sounds {
		for col := range beats {
			if b[row][col] {
				str += "X"
			} else {
				str += "."
			}
		}
		str += "\n"
	}
	return str
}

func New() *Beats {
	return &Beats{
		beats:    BeatState{},
		coordsCh: make(chan programs.Coord, programs.ChannelBuffer),
		// beatsCh:  make(chan BeatState, programs.ChannelBuffer),
		playCh: make(chan string, programs.ChannelBuffer),
		ledsCh: make(chan leds.LEDs, programs.ChannelBuffer),
	}
}

func (b *Beats) Press(press programs.Coord) {
	b.coordsCh <- press
}

func (b *Beats) Play() <-chan string {
	return b.playCh
}
func (b *Beats) Render() <-chan leds.LEDs {
	return b.ledsCh
}

// func (b *Beats) State() <-chan BeatState {
// 	return b.beatsCh
// }

// TODO: decay
// TODO: tempo changes?
func (b *Beats) Run() {
	for coords := range b.coordsCh {
		sound := coords.Row
		beat := coords.Col

		b.beats[sound][beat] = !b.beats[sound][beat]

		// TODO: translate BeatState to LEDs
		// b.beatsCh <- b.beats
	}
}

func (b *Beats) Yield() <-chan struct{} {
	return nil
}
