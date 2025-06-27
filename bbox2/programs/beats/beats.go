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

		// TODO: move to a local var
		timers timers

		log *log.Entry
	}

	interval struct {
		ticksPerBeat int
		ticks        int // TODO: what to do with this?
	}
)

const (
	defaultBPM          = 120
	minBPM              = 30
	maxBPM              = 480
	soundCount          = program.Rows
	beatCount           = program.Cols
	defaultTicksPerBeat = 10
	defaultTicks        = beatCount * defaultTicksPerBeat

	// if 33% of beats are active, yield to the next program
	beatLimit = soundCount * beatCount / 3

	// test
	// decay             = 2 * time.Second
	// keepAliveInterval = 5 * time.Second
	tempoDecay = 5 * time.Second

	// prod
	decay             = 3 * time.Minute
	keepAliveInterval = 14 * time.Minute
	// tempoDecay = 3 * time.Minute
)

var (
	tempoUp   = program.Coord{Row: 0, Col: program.Cols - 1}
	tempoDown = program.Coord{Row: 1, Col: program.Cols - 1}
)

func NewProgram(ctx context.Context) program.Program {
	log := log.WithField("program", "beats")
	log.Debug("NewProgram")

	ctx, cancel := context.WithCancel(ctx)
	b := &beats{
		ctx:    ctx,
		cancel: cancel,

		in:     make(chan program.Coord, program.ChannelBuffer),
		play:   make(chan string, program.ChannelBuffer),
		render: make(chan leds.State, program.ChannelBuffer),
		yield:  make(chan struct{}, program.ChannelBuffer),

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

	sounds := []string{
		"hihat-808.wav",
		"kick-classic.wav",
		"perc-808.wav",
		"tom-808.wav",
	}

	beatState := state{}
	iv := interval{
		ticksPerBeat: defaultTicksPerBeat,
		ticks:        defaultTicks,
	}
	bpm := defaultBPM
	bpmCh := make(chan int, program.ChannelBuffer)

	keepAlive := time.NewTimer(keepAliveInterval)
	if !keepAlive.Stop() {
		<-keepAlive.C
	}
	defer func() {
		if !keepAlive.Stop() {
			select {
			case <-keepAlive.C:
			default:
			}
		}
	}()

	tempoReset := time.NewTimer(tempoDecay)
	if !tempoReset.Stop() {
		<-tempoReset.C
	}
	defer func() {
		if !tempoReset.Stop() {
			select {
			case <-tempoReset.C:
			default:
			}
		}
	}()

	ticker := time.NewTicker(getInterval(bpm, iv.ticksPerBeat))
	defer ticker.Stop()
	tick := 0
	tickTime := time.Now()

	lastPress := program.Coord{Row: -1, Col: -1}

	// starter beat
	b.in <- program.Coord{Row: 1, Col: 0}
	b.in <- program.Coord{Row: 1, Col: 8}

	for {
		select {
		case <-b.ctx.Done():
			return

		// beat loop
		case <-ticker.C:
			tick = (tick + 1) % iv.ticks

			ledsState := leds.State{}
			for i := range 30 {
				ledsState.Set(0, i, leds.Black)
			}
			ledsState.Set(0, tick%30, leds.Mint)
			for _, beat := range beatState {
				for j, active := range beat {
					if active {
						// set the beat LED to red
						ledsState.Set(0, j, leds.Red)
					}
				}
			}
			b.render <- ledsState

			// for each beat type
			if tick%iv.ticksPerBeat == 0 {
				for i, beat := range beatState {
					if beat[tick/iv.ticksPerBeat] {
						// initiate playback
						b.play <- sounds[i]
					}
				}
			}

			t := time.Now()
			b.log.Tracef("BPM:__%+v_", bpm)
			b.log.Tracef("int:__%+v_", getInterval(bpm, iv.ticksPerBeat))
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
			disabling := beatState[press.Row][press.Col]
			beatState[press.Row][press.Col] = !disabling

			b.log.Debugf("Updated beat state:\n%s", beatState)

			if disabling {
				// disabling a beat

				if beatState.allOff() {
					if !keepAlive.Stop() {
						select {
						case <-keepAlive.C:
						default:
						}
					}
					keepAlive.Reset(keepAliveInterval)
				}
			} else {
				// enabling a beat

				if beatState.activeButtons() >= beatLimit {
					b.log.Debugf("Beat limit reached (%d active buttons), yielding...", beatState.activeButtons())
					b.yield <- struct{}{}
					return
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
				if !keepAlive.Stop() {
					select {
					case <-keepAlive.C:
					default:
					}
				}
			}

			// check for tempo changes
			increasingTempo := lastPress == tempoUp && press == tempoUp && bpm < maxBPM
			decreasingTempo := lastPress == tempoDown && press == tempoDown && bpm > minBPM
			if increasingTempo || decreasingTempo {
				if increasingTempo {
					bpmCh <- bpm + 4
				}
				if decreasingTempo {
					bpmCh <- bpm - 4
				}
			}
			lastPress = press

		case newBPM := <-bpmCh:
			b.log.Debugf("BPM changed to %d", newBPM)

			bpm = newBPM

			// BPM: 30 -> 60 -> 120 -> 240 -> 480.0
			// TPB: 40 -> 20 ->  10 ->   5 ->   2.5
			iv.ticksPerBeat = 1200 / bpm
			iv.ticks = beatCount * iv.ticksPerBeat

			// for _, ch := range l.intervalCh {
			// 	ch <- l.iv
			// }

			// reset the tempo after a decay period
			if !tempoReset.Stop() {
				select {
				case <-tempoReset.C:
				default:
				}
			}
			if newBPM != defaultBPM {
				tempoReset.Reset(tempoDecay)
			}

			ticker.Stop()
			ticker = time.NewTicker(getInterval(bpm, iv.ticksPerBeat))
			defer ticker.Stop()

		case <-keepAlive.C:
			select {
			case b.in <- program.Coord{Row: 1, Col: 0}:
			default:
			}

		case <-tempoReset.C:
			bpmCh <- defaultBPM
		}
	}
}

func (b *beats) Yield() <-chan struct{} {
	return b.yield
}

func getInterval(bpm int, ticksPerBeat int) time.Duration {
	return 60 * time.Second / time.Duration(bpm) / (beatCount / 4) / time.Duration(ticksPerBeat) // 4 beats per interval
}
