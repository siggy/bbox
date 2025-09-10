package song

import (
	"context"
	"sync"
	"time"

	"github.com/siggy/bbox/pkg/equalizer"
	"github.com/siggy/bbox/pkg/leds"
	"github.com/siggy/bbox/pkg/program"
	"github.com/siggy/bbox/pkg/rows"
	log "github.com/sirupsen/logrus"
)

type (
	songProgram struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		eq     chan equalizer.DisplayData
		play   chan string
		render chan leds.State
		yield  chan struct{}

		song   string
		length time.Duration

		log *log.Entry
	}
)

const (
	// practically disable color rotation
	ticksPerColorRotation = 10000000 // 60
	pressThreshold        = 10
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
			eq:     make(chan equalizer.DisplayData, program.ChannelBuffer),
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

func (s *songProgram) EQ(displayData equalizer.DisplayData) {
	s.log.Tracef("EQ: %+v", displayData)

	select {
	case <-s.ctx.Done():
		return
	default:
	}

	// enqueue input non-blockingly
	select {
	case s.eq <- displayData:
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

	bands := initBands()

	colorPos := 0
	colorTicks := 0

	s.play <- s.song

	presses := 0

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
			presses++

			if presses >= pressThreshold {
				s.yield <- struct{}{}
				return
			}

		case displayData := <-s.eq:
			s.log.Tracef("Processing EQ: %+v", displayData)

			colors := equalizer.Colorize(displayData)

			s.log.Tracef("Rendering colors: %+v", colors)

			ledsState := leds.State{}

			for row, eqColors := range colors {
				rotatedRow := (row + colorPos) % equalizer.HistorySize

				s.log.Tracef("eqColors[%d]: %+v", rotatedRow, eqColors)
				for i, pixel := range rows.FlatRows[rotatedRow].Pixels {
					color := eqColors[bands[rotatedRow][i]]
					ledsState.Set(pixel.Strip, pixel.Pixel, color)
				}
			}

			s.render <- ledsState

			colorTicks++
			if colorTicks == ticksPerColorRotation {
				colorPos = (colorPos - 1 + equalizer.HistorySize) % equalizer.HistorySize // Cycle through colors
				colorTicks = 0
			}
		}
	}
}

// initBands maps each pixel to its closest button.
func initBands() [program.Rows][]int {
	button := 0

	bands := [program.Rows][]int{}

	for row := range program.Rows {
		bands[row] = make([]int, len(rows.FlatRows[row].Pixels))

		for i := range rows.FlatRows[row].Pixels {
			// figure out which button we are closest to
			prevButtonIndex := rows.FlatRows[row].Buttons[(button-1+program.Cols)%program.Cols]
			nextButtonIndex := rows.FlatRows[row].Buttons[button]

			distanceToPrev := i - prevButtonIndex
			if distanceToPrev < 0 {
				distanceToPrev = len(rows.FlatRows[row].Pixels) - prevButtonIndex + i
			}
			distanceToNext := nextButtonIndex - i
			if distanceToNext < 0 {
				distanceToNext = len(rows.FlatRows[row].Pixels) - i + nextButtonIndex
			}

			if distanceToPrev < 0 || distanceToNext < 0 {
				log.Errorf("Bad distances: row %d button %d i: %d prev: %d next: %d distToPrev: %d distToNext: %d",
					row, button, i, prevButtonIndex, nextButtonIndex, distanceToPrev, distanceToNext)
			}

			bands[row][i] = button

			if distanceToPrev < distanceToNext {
				bands[row][i] = (button - 1 + program.Cols) % program.Cols
			}

			if distanceToNext == 0 {
				button = (button + 1) % program.Cols
			}
		}
	}

	return bands
}
