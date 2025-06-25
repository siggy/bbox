package beats

import (
	"time"
)

const (
	defaultBPM          = 120
	minBPM              = 30
	maxBPM              = 480
	soundCount          = 4
	beatCount           = 16
	defaultTicksPerBeat = 10
	defaultTicks        = beatCount * defaultTicksPerBeat

	tempoDecay = 3 * time.Minute
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
