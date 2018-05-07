package renderer

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox/color"
)

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

// utility functions

func ledToTerm(in uint32) termbox.Attribute {
	return termbox.Attribute(in * 6 / 256)
}

// r, g, b, w => (fg(w), bg(r,g,b))
func ledColorToTermColor(value uint32) (termbox.Attribute, termbox.Attribute) {
	r, g, b, w := color.ParseColor(value)

	// these constants assume termbox.Output256
	fg := 0xe9 + termbox.Attribute(w*24/256)
	bg := 0x11 + ledToTerm(r)*36 + ledToTerm(g)*6 + ledToTerm(b)

	return fg, bg
}
