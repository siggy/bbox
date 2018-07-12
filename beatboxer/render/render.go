package render

import (
	"fmt"

	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/beatboxer/color"
)

// TODO: these need to be in an output-agnostic module
const (
	ROWS          = 4
	COLUMNS       = 16
	COLS_PER_BEAT = 10
)

type Transition struct {
	Color    uint32
	Location float64 // [0,1]
	Length   float64 // [0,1]
}

type RenderState struct {
	LEDs        [ROWS][COLUMNS]uint32
	Transitions [ROWS][COLUMNS]Transition
}

type Renderer interface {
	Init(
		freq int,
		gpioPin1 int, ledCount1 int, brightness1 int,
		gpioPin2 int, ledCount2 int, brightness2 int,
	) error

	Fini()
	Render() error
	Wait() error
	SetLed(channel int, index int, value uint32)
	Clear()
	SetBitmap(channel int, a []uint32)
}

// TODO: dry up
func TBprint(x, y int, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

func Render(renderState RenderState) {

	TBprint(0, 10, fmt.Sprintf("______RenderState:_%+v______", renderState))

	for row := 0; row < ROWS; row++ {
		for col := 0; col < COLUMNS; col++ {
			// clear everything
			for i := 0; i < COLS_PER_BEAT; i++ {
				termbox.SetCell(col*COLS_PER_BEAT+i, row, ' ', termbox.ColorBlack, termbox.ColorBlack)
			}

			fgColor := termbox.ColorBlack
			rune := ' '
			if renderState.LEDs[row][col] == color.Make(0, 0, 0, 127) {
				fgColor = termbox.ColorWhite
				rune = 'X'
			} else if renderState.LEDs[row][col] == color.Make(127, 0, 0, 0) {
				fgColor = termbox.ColorRed
				rune = 'O'
			} else if renderState.LEDs[row][col] == color.Make(250, 143, 94, 0) {
				// ceottk
				fgColor = termbox.ColorYellow
				rune = 'X'
			}

			if renderState.Transitions[row][col].Color == color.Make(0, 0, 0, 127) {
				location := int(COLS_PER_BEAT * renderState.Transitions[row][col].Location)
				termbox.SetCell(col*COLS_PER_BEAT+location, row, 'X', termbox.ColorBlack, termbox.ColorWhite)
			}

			termbox.SetCell(col*COLS_PER_BEAT, row, rune, termbox.ColorBlack, fgColor)
		}
	}

	termbox.Flush()

	// TODO: map to actual LEDs
}
