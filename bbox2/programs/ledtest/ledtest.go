package ledtest

import (
	"context"
	"sync"

	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	log "github.com/sirupsen/logrus"
)

type (
	ledTest struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		render chan leds.State
		yield  chan struct{}

		log *log.Entry
	}
)

func New(ctx context.Context) program.Program {
	log := log.WithField("program", "ledTest")
	log.Debug("New")

	ctx, cancel := context.WithCancel(ctx)
	l := &ledTest{
		ctx:    ctx,
		cancel: cancel,

		in:     make(chan program.Coord, program.ChannelBuffer),
		render: make(chan leds.State, program.ChannelBuffer),
		yield:  make(chan struct{}, program.ChannelBuffer),

		log: log,
	}

	l.wg.Add(1)
	go l.run()

	return l
}

func (l *ledTest) Name() string {
	return "ledTest"
}

func (l *ledTest) Close() {
	l.cancel()
	l.wg.Wait()
	close(l.render)
}

func (l *ledTest) Press(press program.Coord) {
	l.log.Debugf("Press: %+v", press)

	select {
	case <-l.ctx.Done():
		return
	default:
	}
	// enqueue input non-blockingly
	select {
	case l.in <- press:
	default:
	}
}

func (l *ledTest) Play() <-chan string {
	return nil
}
func (l *ledTest) Render() <-chan leds.State {
	return l.render
}

func (l *ledTest) Yield() <-chan struct{} {
	return l.yield
}

func (l *ledTest) run() {
	defer l.wg.Done()

	lState := leds.State{}
	lState.Set(0, 0, leds.Red) // first pixel lit

	for {
		select {
		case <-l.ctx.Done():
			return

		case press := <-l.in:
			l.log.Debugf("Processing press: %+v", press)

			lState.Set(press.Row, press.Col, leds.Red)

			l.render <- lState
		}
	}
}
