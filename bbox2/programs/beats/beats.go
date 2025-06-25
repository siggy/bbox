package beats

import (
	"context"
	"sync"
	"time"

	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	log "github.com/sirupsen/logrus"
)

const (
	sounds    = 4
	beatCount = 16
)

type (
	beatState [sounds][beatCount]bool

	beats struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		play   chan string
		render chan leds.State
		yield  chan struct{}

		state beatState

		log *log.Entry
	}
)

func (b beatState) String() string {
	var str string
	for row := range sounds {
		for col := range beatCount {
			if b[row][col] {
				str += "X"
			} else {
				str += "."
			}
		}
		str += "\n"
	}
	return str
}

func NewProgram(ctx context.Context) program.Program {
	ctx, cancel := context.WithCancel(ctx)
	b := &beats{
		ctx:    ctx,
		cancel: cancel,

		state: beatState{},

		in:     make(chan program.Coord, program.ChannelBuffer),
		play:   make(chan string, program.ChannelBuffer),
		render: make(chan leds.State, program.ChannelBuffer),
		yield:  make(chan struct{}, program.ChannelBuffer),

		log: log.WithField("program", "beats"),
	}
	b.wg.Add(1)
	go b.run() // fan-in keyboard, timers, etc
	return b
}

func (b *beats) Close() {
	b.cancel()
	b.wg.Wait()
	close(b.play)
	close(b.render)
}

func (b *beats) Press(press program.Coord) {
	b.log.Debugf("Press: %+v", press)

	select {
	case <-b.ctx.Done():
		return
	default:
	}
	// enqueue input non-blockingly
	select {
	case b.in <- press:
	default:
	}
	// if 'n' pressed, signal yield once
	// if lower right is pressed 5 times, yield
	// if press.Rune == 'n' {
	// 	select {
	// 	case b.yield <- struct{}{}:
	// 	default:
	// 	}
	// }
}

func (b *beats) Play() <-chan string {
	return b.play
}
func (b *beats) Render() <-chan leds.State {
	return b.render
}

// TODO: decay
// TODO: tempo changes?
func (b *beats) run() {
	// for coords := range b.inCoords {
	// 	sound := coords.Row
	// 	beat := coords.Col

	// 	b.beats[sound][beat] = !b.beats[sound][beat]

	// 	// TODO: translate beatState to LEDs
	// 	// b.beatsCh <- b.beats
	// }

	defer b.wg.Done()

	// 120 BPM → 500ms per step
	ticker := time.NewTicker(time.Minute / 120)
	defer ticker.Stop()

	// pattern: kick, rest, snare, rest
	sounds := []string{"kick.wav", "", "snare.wav", ""}

	l := leds.State{}
	l.Set(0, 0, leds.Red) // first pixel lit

	step := 0
	for {
		select {
		case <-b.ctx.Done():
			return

		case <-ticker.C:
			if sounds[step] != "" {
				b.play <- sounds[step]
			}
			b.render <- l
			step = (step + 1) % len(sounds)

		case <-b.in:
			// ignore other presses here
		}
	}

}

func (b *beats) Yield() <-chan struct{} {
	b.log.Debug("Yield")

	return b.yield
}
