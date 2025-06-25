package beats

import (
	"time"

	"github.com/siggy/bbox/bbox2/program"
)

const (
	defaultBPM          = 120
	minBPM              = 30
	maxBPM              = 480
	soundCount          = program.Rows
	beatCount           = program.Cols
	defaultTicksPerBeat = 10
	defaultTicks        = beatCount * defaultTicksPerBeat

	tempoDecay = 3 * time.Minute

	// if 33% of beats are active, yield to the next program
	beatLimit = soundCount * beatCount / 3
)

type interval struct {
	ticksPerBeat int
	ticks        int
}

var sounds = []string{
	"hihat-808.wav",
	"kick-classic.wav",
	"perc-808.wav",
	"tom-808.wav",
}
