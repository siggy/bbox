package beatboxer

import (
	"time"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/render"
)

// ProgramOutput defines the interface Beatboxer programs may send output information
type ProgramOutput interface {
	Play(uint32, string)
	Render(uint32, render.State)
	Yield(uint32)
}

// Program defines the interface all Beatboxer programs must satisfy
type Program interface {
	New(uint32, ProgramOutput, map[string]time.Duration) Program

	// input
	Amplitude(float64)
	Keyboard(bbox.Coord)
	Close()
}
