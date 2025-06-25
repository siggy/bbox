package beats

import (
	"time"

	"github.com/siggy/bbox/bbox2/program"
)

type interval struct {
	ticksPerBeat int
	ticks        int // TODO: what to do with this?
}

const (
	defaultBPM          = 120
	minBPM              = 30
	maxBPM              = 480
	soundCount          = program.Rows
	beatCount           = program.Cols
	defaultTicksPerBeat = 10
	defaultTicks        = beatCount * defaultTicksPerBeat

	// if 33% of beats are active, yield to the next program
	beatLimit = soundCount * beatCount / 3

	// test
	// decay      = 2 * time.Second
	// keepAlive  = 5 * time.Second
	tempoDecay = 5 * time.Second

	// prod
	decay     = 3 * time.Minute
	keepAlive = 14 * time.Minute
	// tempoDecay = 3 * time.Minute
)

var (
	tempoUp   = program.Coord{Row: 0, Col: program.Cols - 1}
	tempoDown = program.Coord{Row: 1, Col: program.Cols - 1}
)
