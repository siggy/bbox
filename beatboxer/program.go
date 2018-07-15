package beatboxer

import (
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/render"
)

// Program defines the interface all Beatboxer programs must satisfy
type Program interface {
	New() Program

	// input
	Amplitude() chan<- float64
	Keyboard() chan<- bbox.Coord
	Close() chan<- struct{}

	// output
	Play() <-chan string
	Render() <-chan render.RenderState
	Yield() <-chan struct{}
}
