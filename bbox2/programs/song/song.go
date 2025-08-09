package song

import (
	"context"
	"sync"
	"time"

	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	log "github.com/sirupsen/logrus"
)

type (
	songProgram struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		play   chan string
		render chan leds.State
		yield  chan struct{}

		song   string
		length time.Duration

		log *log.Entry
	}
)

func New(song string, length time.Duration) program.ProgramFactory {
	return func(ctx context.Context) program.Program {
		log := log.WithFields(log.Fields{"program": "song", "song": song, "length": length})
		log.Debug("New")

		ctx, cancel := context.WithCancel(ctx)
		l := &songProgram{
			ctx:    ctx,
			cancel: cancel,

			in:     make(chan program.Coord, program.ChannelBuffer),
			play:   make(chan string, program.ChannelBuffer),
			render: make(chan leds.State, program.ChannelBuffer),
			yield:  make(chan struct{}, program.ChannelBuffer),

			song:   song,
			length: length,

			log: log,
		}

		l.wg.Add(1)
		go l.run()

		return l
	}
}

func (s *songProgram) Name() string {
	return s.song
}

func (s *songProgram) Close() {
	s.cancel()
	s.wg.Wait()
	close(s.play)
	close(s.render)
}

func (s *songProgram) Press(press program.Coord) {
	s.log.Debugf("Press: %+v", press)

	select {
	case <-s.ctx.Done():
		return
	default:
	}
	// enqueue input non-blockingly
	select {
	case s.in <- press:
	default:
	}
}

func (s *songProgram) Play() <-chan string {
	return nil
}
func (s *songProgram) PlayWithEQ() <-chan string {
	return s.play
}
func (s *songProgram) Render() <-chan leds.State {
	return s.render
}
func (s *songProgram) Yield() <-chan struct{} {
	return s.yield
}

func (s *songProgram) run() {
	defer s.wg.Done()

	lState := leds.State{}
	lState.Set(0, 0, leds.Red) // first pixel lit

	s.play <- s.song

	// song duration
	timer := time.NewTimer(s.length)

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-timer.C:
			s.log.Debug("Timer expired, stopping program")
			s.yield <- struct{}{}
			return

		case press := <-s.in:
			s.log.Debugf("Processing press: %+v", press)
		}
	}
}
