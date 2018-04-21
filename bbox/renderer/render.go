package renderer

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
