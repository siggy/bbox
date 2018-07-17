package render

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

type State struct {
	LEDs        [ROWS][COLUMNS]uint32
	Transitions [ROWS][COLUMNS]Transition
}

type Renderer interface {
	Render(state State)
}

// TODO: move this to somewhere more LED specific ?

type LedRenderer interface {
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
