package bbox

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio"
)

type Gpio struct {
	debounce  chan struct{}
	debounced bool
	press     chan<- struct{} // single button press, gpio->emitter
	wg        *sync.WaitGroup
}

func InitGpio(wg *sync.WaitGroup, press chan<- struct{}) *Gpio {
	wg.Add(1)

	return &Gpio{
		debounce:  make(chan struct{}),
		debounced: true,
		press:     press,
		wg:        wg,
	}
}

func (g *Gpio) Run() {
	defer g.wg.Done()

	if err := rpio.Open(); err != nil {
		panic(err)
	}
	defer rpio.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

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
				time.AfterFunc(time.Second, func() {
					g.debounce <- struct{}{}
				})
			}
		case <-g.debounce:
			g.debounced = true
		case <-sig:
			return
		default:
		}
	}
}
