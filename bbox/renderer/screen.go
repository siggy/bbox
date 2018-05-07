package renderer

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/color"
)

// WS2811 satisfies the Render interface, backed by the ws2811 LED library
type Screen struct{}

const INFO_ROW = 30
const CHANNEL_MULTIPLIER = 15

var (
	runes = []rune{' ', '░', '▒', '▓', '█'}

	// globals to approximate ws2811 interface / behavior
	gLedCount1   int
	gLedCount2   int
	gBrightness1 int
	gBrightness2 int
)

func (w Screen) Init(
	freq int,
	gpioPin1 int, ledCount1 int, brightness1 int,
	gpioPin2 int, ledCount2 int, brightness2 int,
) error {
	gLedCount1 = ledCount1
	gLedCount2 = ledCount2
	gBrightness1 = brightness1
	gBrightness2 = brightness2

	bbox.Tbprint(0, INFO_ROW-2,
		"Init(%7d, %3d, %3d, %3d, %3d, %3d, %3d)",
		freq,
		gpioPin1, ledCount1, brightness1,
		gpioPin2, ledCount2, brightness2,
	)
	termbox.Flush()
	return nil
}

func (s Screen) Fini() {
	bbox.Tbprint(0, INFO_ROW, "Fini")
	termbox.Flush()
}

func (s Screen) Render() error {
	bbox.Tbprint(0, INFO_ROW, "Render")
	termbox.Flush()
	return nil
}

func (s Screen) Wait() error {
	bbox.Tbprint(0, INFO_ROW, "Wait")
	termbox.Flush()
	return nil
}

func (s Screen) SetLed(channel int, index int, value uint32) {
	w, _ := termbox.Size()
	y := index / w
	x := index - y*w

	bbox.Tbprint(0, INFO_ROW-1, "SetLed(%2d, (%3d, %3d), %20s)", channel, x, y, color.ColorStr(value))
	fg, bg := ledColorToTermColor(value)
	termbox.SetCell(x, y+channel*CHANNEL_MULTIPLIER, '░', fg, bg)
	termbox.Flush()
}

func (s Screen) Clear() {
	bbox.Tbprint(0, INFO_ROW, "Clear")
	for y := 0; y < 2; y++ {
		for x := 0; x < gLedCount1; x++ {
			// termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorBlack)
		}
	}
	termbox.Clear(termbox.ColorBlack, termbox.ColorBlack)
}

func (s Screen) SetBitmap(channel int, a []uint32) {
	bbox.Tbprint(INFO_ROW, 0, "SetBitmap(%d, %+v)", channel, a)
	termbox.Flush()
}
