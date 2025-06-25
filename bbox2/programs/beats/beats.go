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
	beats struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		in     chan program.Coord
		play   chan string
		render chan leds.State
		yield  chan struct{}

		bpmCh chan int

		state state
		iv    interval
		bpm   int

		timers     timers
		keepAlive  *time.Timer // ensure at least one beat is sent periodically to keep speaker alive
		tempoDecay *time.Timer // decay timer for tempo changes

		log *log.Entry
	}
)

func NewProgram(ctx context.Context) program.Program {
	log := log.WithField("program", "beats")
	log.Debug("NewProgram ")

	ctx, cancel := context.WithCancel(ctx)
	b := &beats{
		ctx:    ctx,
		cancel: cancel,

		state: state{},
		iv: interval{
			ticksPerBeat: defaultTicksPerBeat,
			ticks:        defaultTicks,
		},
		bpm: defaultBPM,

		in:     make(chan program.Coord, program.ChannelBuffer),
		play:   make(chan string, program.ChannelBuffer),
		render: make(chan leds.State, program.ChannelBuffer),
		yield:  make(chan struct{}, program.ChannelBuffer),

		bpmCh: make(chan int, program.ChannelBuffer),

		log: log,
	}

	b.wg.Add(1)
	go b.run()

	return b
}

func (b *beats) Name() string {
	return "beats"
}

func (b *beats) Close() {
	b.cancel()
	b.wg.Wait()
	close(b.play)
	close(b.render)

	if b.keepAlive != nil {
		b.keepAlive.Stop()
	}

	if b.tempoDecay != nil {
		b.tempoDecay.Stop()
	}

	for _, arr := range b.timers {
		for _, t := range arr {
			if t != nil {
				t.Stop()
			}
		}
	}
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
}

func (b *beats) Play() <-chan string {
	return b.play
}
func (b *beats) Render() <-chan leds.State {
	return b.render
}

func (b *beats) run() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.getInterval())
	defer ticker.Stop()
	tick := 0
	tickTime := time.Now()

	sounds := []string{
		"hihat-808.wav",
		"kick-classic.wav",
		"perc-808.wav",
		"tom-808.wav",
	}

	l := leds.State{}
	l.Set(0, 0, leds.Red) // first pixel lit

	// starter beat
	b.in <- program.Coord{Row: 1, Col: 0}
	b.in <- program.Coord{Row: 1, Col: 8}

	lastPress := program.Coord{Row: -1, Col: -1}

	for {
		select {
		case <-b.ctx.Done():
			return

		// beat loop
		case <-ticker.C:
			tick = (tick + 1) % b.iv.ticks

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
			b.log.Tracef("int:__%+v_", b.getInterval())
			b.log.Tracef("time:_%+v_", t.Sub(tickTime))
			b.log.Tracef("tick:_%+v_", tick)
			tickTime = t

		case press := <-b.in:
			b.log.Debugf("Processing press: %+v", press)

			if press.Row < 0 || press.Row >= soundCount || press.Col < 0 || press.Col >= beatCount {
				b.log.Warnf("Invalid press coordinates: %+v", press)
				continue
			}

			// disable decay timer
			if b.timers[press.Row][press.Col] != nil {
				b.timers[press.Row][press.Col].Stop()
			}

			// toggle the beat state
			disabling := b.state[press.Row][press.Col]
			b.state[press.Row][press.Col] = !disabling

			increasingTempo := lastPress == tempoUp && press == tempoUp && b.bpm < maxBPM
			decreasingTempo := lastPress == tempoDown && press == tempoDown && b.bpm > minBPM
			if increasingTempo || decreasingTempo {
				if increasingTempo {
					b.bpmCh <- b.bpm + 4
				}
				if decreasingTempo {
					b.bpmCh <- b.bpm - 4
				}
			}
			lastPress = press

			if disabling {
				// disabling a beat
				if b.state.allOff() {
					b.keepAlive = time.AfterFunc(keepAlive, func() {
						// enable a beat to keep the speaker alive
						b.log.Debug("Keep alive timer expired")

						b.in <- program.Coord{Row: 1, Col: 0}
					})
				}
			} else {
				// enabling a beat
				if b.state.activeButtons() >= beatLimit {
					b.log.Debugf("Beat limit reached (%d active buttons), yielding...", b.state.activeButtons())
					select {
					case b.yield <- struct{}{}:
					default:
						b.log.Warn("Yield channel is full")
					}
				}

				// set a decay timer
				b.timers[press.Row][press.Col] = time.AfterFunc(decay, func() {
					b.log.Debugf("Decay timer expired for press: %+v", press)
					b.in <- program.Coord{
						Row: press.Row,
						Col: press.Col,
					}
				})

				// we've enabled a beat, kill keepAlive
				if b.keepAlive != nil {
					b.keepAlive.Stop()
				}
			}

			b.log.Debugf("Updated beat state:\n%s", b.state)
		case bpm := <-b.bpmCh:
			b.log.Debugf("BPM changed to %d", bpm)

			b.bpm = bpm

			// set a decay timer
			if b.tempoDecay != nil {
				b.tempoDecay.Stop()
			}
			b.tempoDecay = time.AfterFunc(tempoDecay, func() {
				b.bpmCh <- defaultBPM
			})

			ticker.Stop()
			ticker = time.NewTicker(b.getInterval())
			defer ticker.Stop()
		}
	}
}

func (b *beats) Yield() <-chan struct{} {
	return b.yield
}

func (b *beats) getInterval() time.Duration {
	return 60 * time.Second / time.Duration(b.bpm) / (beatCount / 4) / time.Duration(b.iv.ticksPerBeat) // 4 beats per interval
}
