package devil

import (
	"context"
	"sync"
	"time"

	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	log "github.com/sirupsen/logrus"
)

type (
	devil struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		play   chan string
		render chan leds.State
		yield  chan struct{}

		log *log.Entry
	}
)

func New(ctx context.Context) program.Program {
	log := log.WithField("program", "devil")
	log.Debug("New")

	ctx, cancel := context.WithCancel(ctx)
	d := &devil{
		ctx:    ctx,
		cancel: cancel,

		in:     make(chan program.Coord, program.ChannelBuffer),
		play:   make(chan string, program.ChannelBuffer),
		render: make(chan leds.State, program.ChannelBuffer),
		yield:  make(chan struct{}, program.ChannelBuffer),

		log: log,
	}

	d.wg.Add(1)
	go d.run()

	return d
}

func (d *devil) Name() string {
	return "devil"
}

func (d *devil) Close() {
	d.cancel()
	d.wg.Wait()
	close(d.play)
	close(d.render)
}

func (d *devil) Press(press program.Coord) {
	d.log.Debugf("Press: %+v", press)

	select {
	case <-d.ctx.Done():
		return
	default:
	}
	// enqueue input non-blockingly
	select {
	case d.in <- press:
	default:
	}
}

func (d *devil) EQ([]float64) {}

func (d *devil) Play() <-chan string {
	return d.play
}
func (d *devil) Render() <-chan leds.State {
	return d.render
}
func (d *devil) Yield() <-chan struct{} {
	return d.yield
}

func (d *devil) run() {
	defer d.wg.Done()

	lState := leds.State{}
	lState.Set(0, 0, leds.Red) // first pixel lit

	d.play <- "runnin_with_the_devid.wav"

	// song duration
	timer := time.NewTimer(time.Second * 215)

	for {
		select {
		case <-d.ctx.Done():
			return

		case <-timer.C:
			d.log.Debug("Timer expired, stopping program")
			d.yield <- struct{}{}
			return

		case press := <-d.in:
			d.log.Debugf("Processing press: %+v", press)
		}
	}
}
