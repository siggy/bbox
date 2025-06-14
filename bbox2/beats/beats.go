package beats

import "github.com/siggy/bbox/bbox2/keyboard"

const (
	sounds = 4
	beats  = 16
)

type (
	BeatState [sounds][beats]bool

	Beats struct {
		beats    BeatState
		coordsCh chan keyboard.Coord
		beatsCh  chan BeatState
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
		coordsCh: make(chan keyboard.Coord, 100),
		beatsCh:  make(chan BeatState, 100),
	}
}

func (b *Beats) Press(press keyboard.Coord) {
	b.coordsCh <- press
}

func (b *Beats) State() <-chan BeatState {
	return b.beatsCh
}

// TODO: decay
// TODO: tempo changes?
func (b *Beats) Run() {
	for coords := range b.coordsCh {
		sound := coords.Row
		beat := coords.Col

		b.beats[sound][beat] = !b.beats[sound][beat]

		b.beatsCh <- b.beats
	}
}
