package beats

import log "github.com/sirupsen/logrus"

const (
	sounds = 4
	beats  = 16
)

type (
	BeatState [sounds][beats]bool

	Beats struct {
		keymaps map[rune]*Coord
		beats   BeatState
		keyCh   chan rune
		beatsCh chan BeatState
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

func New(keymaps map[rune]*Coord) *Beats {
	return &Beats{
		keymaps: keymaps,
		beats:   BeatState{},
		keyCh:   make(chan rune, 100),
		beatsCh: make(chan BeatState, 100),
	}
}

func (b *Beats) Press(press rune) {
	b.keyCh <- press
}

func (b *Beats) State() <-chan BeatState {
	return b.beatsCh
}

// TODO: decay
func (b *Beats) Run() {
	for press := range b.keyCh {
		coords, ok := b.keymaps[press]
		if !ok {
			log.Warnf("No coordinates for key %q", press)
			continue
		}

		sound := coords.Row
		beat := coords.Col

		b.beats[sound][beat] = !b.beats[sound][beat]

		b.beatsCh <- b.beats
	}
}
