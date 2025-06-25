package beats

import (
	"context"
	"sync"
	"time"

	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/bbox/bbox2/program"
	log "github.com/sirupsen/logrus"
)

type (
	beatState [soundCount][beatCount]bool

	beats struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		play   chan string
		render chan leds.State
		yield  chan struct{}

		state beatState
		iv    interval
		bpm   int

		log *log.Entry
	}
)

func (b beatState) String() string {
	var str string
	for row := range sounds {
		for col := range b[row] {
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
	log := log.WithField("program", "beats")
	log.Debug("NewProgram ")

	ctx, cancel := context.WithCancel(ctx)
	b := &beats{
		ctx:    ctx,
		cancel: cancel,

		state: beatState{},
		iv: interval{
			ticksPerBeat: defaultTicksPerBeat,
			ticks:        defaultTicks,
		},
		bpm: defaultBPM,

		// in:     make(chan program.Coord, program.ChannelBuffer),
		// play:   make(chan string, program.ChannelBuffer),
		// render: make(chan leds.State, program.ChannelBuffer),
		in:     make(chan program.Coord),
		play:   make(chan string),
		render: make(chan leds.State),

		yield: make(chan struct{}, 1),

		log: log,
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
	b.log.Debugf("Received press: %+v", press)

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

	ticker := time.NewTicker(b.bpmToInterval(b.bpm))
	defer ticker.Stop()
	tick := 0
	tickTime := time.Now()

	// pattern: kick, rest, snare, rest
	sounds := []string{"perc-808.wav", "hihat-808.wav", "kick-classic.wav", "tom-808.wav"}

	l := leds.State{}
	l.Set(0, 0, leds.Red) // first pixel lit

	// step := 0
	for {
		select {
		case <-b.ctx.Done():
			return

		case <-ticker.C: // for every time interval
			// next interval
			tick = (tick + 1) % b.iv.ticks
			// tmp := tick

			// for _, ch := range b.ticks {
			// 	ch <- tmp
			// }

			// for each beat type
			if tick%b.iv.ticksPerBeat == 0 {
				for i, beat := range b.state {
					if beat[tick/b.iv.ticksPerBeat] {
						// initiate playback
						b.play <- sounds[i]
					}
				}
			}

			t := time.Now()
			b.log.Tracef("BPM:__%+v_", b.bpm)
			b.log.Tracef("int:__%+v_", b.bpmToInterval(b.bpm))
			b.log.Tracef("time:_%+v_", t.Sub(tickTime))
			b.log.Tracef("tick:_%+v_", tick)
			tickTime = t
			// if sounds[step] != "" {
			// 	// b.play <- sounds[step]
			// }
			// step = (step + 1) % len(sounds)

			// l.Set(0, step, leds.Red)
			// // b.render <- l

		case press := <-b.in:
			b.log.Debugf("Processing press: %+v", press)

			if press.Row == 3 && press.Col == 15 {
				b.yield <- struct{}{}
				return
			}

			if press.Row < 0 || press.Row >= soundCount || press.Col < 0 || press.Col >= beatCount {
				b.log.Warnf("Invalid press coordinates: %+v", press)
				continue
			}

			// toggle the beat state
			b.state[press.Row][press.Col] = !b.state[press.Row][press.Col]

			b.log.Debugf("Updated beat state:\n%s", b.state)
		}
	}
}

func (b *beats) Yield() <-chan struct{} {
	return b.yield
}

func (b *beats) bpmToInterval(bpm int) time.Duration {
	return 60 * time.Second / time.Duration(bpm) / (beatCount / 4) / time.Duration(b.iv.ticksPerBeat) // 4 beats per interval
}
