package program

import (
	"context"

	"github.com/siggy/bbox/bbox2/equalizer"
	"github.com/siggy/bbox/bbox2/leds"
)

type (
	Coord struct {
		Row int
		Col int
	}

	// Program defines the interface all Beatboxer programs must satisfy
	Program interface {
		Name() string

		// input
		Press(press Coord)
		EQ(equalizer.DisplayData)

		// output
		Play() <-chan string
		PlayWithEQ() <-chan string
		Render() <-chan leds.State
		Yield() <-chan struct{}

		// clean up
		Close()
	}

	ProgramFactory func(ctx context.Context) Program
)

const (
	Rows = 4
	Cols = 16

	// TODO: remove?
	ChannelBuffer = 100
)
