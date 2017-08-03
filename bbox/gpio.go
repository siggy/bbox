package bbox

import (
	"time"

	"github.com/stianeikeland/go-rpio"
)

const (
	DEBOUNCE_TIME = 500 * time.Millisecond
)

type Gpio struct {
	closing   chan struct{}
	debounce  chan struct{}
	debounced bool
	press     chan<- struct{} // single button press, gpio->emitter
}

func InitGpio(press chan<- struct{}) *Gpio {
	return &Gpio{
		debounce:  make(chan struct{}),
		debounced: true,
		closing:   make(chan struct{}),
		press:     press,
	}
}

func (g *Gpio) Run() {
	if err := rpio.Open(); err != nil {
		panic(err)
	}
	defer rpio.Close()

	pin := rpio.Pin(22)
	pin.Input()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	defer close(g.press)

	for {
		select {
		case <-ticker.C:
			res := pin.Read()
			if res == rpio.Low && g.debounced {
				g.debounced = false
				g.press <- struct{}{}
				time.AfterFunc(DEBOUNCE_TIME, func() {
					g.debounce <- struct{}{}
				})
			}
		case <-g.debounce:
			g.debounced = true
		case _, more := <-g.closing:
			if !more {
				return
			}
		default:
		}
	}
}

func (g *Gpio) Close() {
	close(g.closing)
}
