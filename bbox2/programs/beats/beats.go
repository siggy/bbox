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

		log *log.Entry
	}

	timers [soundCount][beatCount]*time.Timer
)

const (
	tickInterval = 50 * time.Millisecond
	defaultBPM   = 120
	minBPM       = 30
	maxBPM       = 480
	soundCount   = program.Rows
	beatCount    = program.Cols

	// if 33% of beats are active, yield to the next program
	beatLimit = soundCount * beatCount / 3

	// test
	// decay             = 2 * time.Second
	// keepAliveInterval = 5 * time.Second
	// tempoDecay        = 5 * time.Second

	// prod
	decay             = 3 * time.Minute
	keepAliveInterval = 14 * time.Minute
	tempoDecay        = 3 * time.Minute
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

	beatIndex := 0
	beatState := state{}

	bpm := defaultBPM
	bpmCh := make(chan int, program.ChannelBuffer)

	beatsPerTick := getBeatsPerTick(bpm)
	var beatAcc float64

	decayCh := make(chan program.Coord, program.ChannelBuffer)

	decayTimers := timers{}
	defer func() {
		for _, arr := range decayTimers {
			for _, t := range arr {
				if t != nil {
					if !t.Stop() {
						select {
						case <-t.C:
						default:
						}
					}
				}
			}
		}
	}()

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

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

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
			beatAcc += beatsPerTick

			ledsState := leds.State{}

			for beatAcc >= 1.0 {
				beatAcc -= 1.0

				// TODO: handle real strips
				// for i := range soundCount {
				for row := range 1 {
					prevIndex := (beatIndex - 1 + beatCount) % beatCount

					if !beatState[row][prevIndex] {
						// do not perturb particular LEDs
						ledsState.Set(row, prevIndex, leds.Black)
					}
					if !beatState[row][beatIndex] {
						// do not perturb particular LEDs
						ledsState.Set(row, beatIndex, leds.Mint)
					}
				}

				for i := range beatState {
					if beatState[i][beatIndex] {
						b.play <- sounds[i]
					}
				}

				beatIndex = (beatIndex + 1) % beatCount
			}

			b.render <- ledsState

			// … do your LED rendering here …

			// ledsState := leds.State{}
			// for i := range 30 {
			// 	ledsState.Set(0, i, leds.Black)
			// }
			// ledsState.Set(0, tick%30, leds.Mint)
			// for _, beat := range beatState {
			// 	for j, active := range beat {
			// 		if active {
			// 			// set the beat LED to red
			// 			ledsState.Set(0, j, leds.Red)
			// 		}
			// 	}
			// }
			// b.render <- ledsState

			t := time.Now()
			b.log.Tracef("BPM:___________%+v_", bpm)
			b.log.Tracef("int:___________%+v_", tickInterval)
			b.log.Tracef("time:__________%+v_", t.Sub(tickTime))
			b.log.Tracef("beatIndex:_____%+v_", beatIndex)
			b.log.Tracef("beatsPerTick:__%+v_", beatsPerTick)
			b.log.Tracef("beatAcc:_______%+v_", beatAcc)
			tickTime = t

		case press := <-b.in:
			b.log.Debugf("Processing press: %+v", press)

			if press.Row < 0 || press.Row >= soundCount || press.Col < 0 || press.Col >= beatCount {
				b.log.Warnf("Invalid press coordinates: %+v", press)
				continue
			}

			// disable decay timer
			if decayTimers[press.Row][press.Col] != nil {
				decayTimers[press.Row][press.Col].Stop()
			}

			// toggle the beat state
			disabling := beatState[press.Row][press.Col]
			beatState[press.Row][press.Col] = !disabling

			b.log.Debugf("Updated beat state:\n%s", beatState)

			color := leds.Red

			if disabling {
				// disabling a beat
				color = leds.Black

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
				decayTimers[press.Row][press.Col] = time.AfterFunc(decay, func() {
					select {
					case decayCh <- press:
					default:
					}
				})

				// we've enabled a beat, kill keepAlive
				if !keepAlive.Stop() {
					select {
					case <-keepAlive.C:
					default:
					}
				}

				// for i := range 30 {
				// 	ledsState.Set(0, i, leds.Black)
				// }
				// ledsState.Set(0, tick%30, leds.Mint)
				// for _, beat := range beatState {
				// 	for j, active := range beat {
				// 		if active {
				// 			// set the beat LED to red
				// 			ledsState.Set(0, j, leds.Red)
				// 		}
				// 	}
				// }
				// b.render <- ledsState
			}

			// TODO: map to actual LED coordinates
			ledsState := leds.State{}
			ledsState.Set(press.Row, press.Col, color)
			b.render <- ledsState

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
			b.log.Debugf("BPM changed from %d to %d", bpm, newBPM)

			bpm = newBPM
			beatsPerTick = getBeatsPerTick(bpm)

			// reset the tempo after a decay period
			if !tempoReset.Stop() {
				select {
				case <-tempoReset.C:
				default:
				}
			}
			if bpm != defaultBPM {
				tempoReset.Reset(tempoDecay)
			}

		case coord := <-decayCh:
			b.log.Debugf("Decay timer expired for press: %+v", coord)
			select {
			case b.in <- coord:
			default:
			}

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

func getBeatsPerTick(bpm int) float64 {
	return beatCount / 4.0 * float64(bpm) / 60.0 * tickInterval.Seconds()
}
