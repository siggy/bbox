package ws2811

import (
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

// WS2811 satisfies the Render interface, backed by the ws2811 LED library
type WS2811 struct{}

func (w WS2811) Init(
	freq int,
	gpioPin1 int, ledCount1 int, brightness1 int,
	gpioPin2 int, ledCount2 int, brightness2 int,
) error {
	return ws2811.Init(
		freq,
		gpioPin1, ledCount1, brightness1,
		gpioPin2, ledCount2, brightness2,
	)
}

func (w WS2811) Fini() {
	ws2811.Fini()
}

func (w WS2811) Render() error {
	return ws2811.Render()
}

func (w WS2811) Wait() error {
	return ws2811.Wait()
}

func (w WS2811) SetLed(channel int, index int, value uint32) {
	ws2811.SetLed(channel, index, value)
}

func (w WS2811) Clear() {
	ws2811.Clear()
}

func (w WS2811) SetBitmap(channel int, a []uint32) {
	ws2811.SetBitmap(channel, a)
}
