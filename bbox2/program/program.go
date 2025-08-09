package program

import (
	"context"

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

// byte(strip), pixel, g, r, b, w
// stripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}

// TODO:
// program interface {
// 	Press() Coord<-
//  Play() <-String
//  Render() <-LEDs
// }
