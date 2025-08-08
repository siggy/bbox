package beats

import (
	"context"
	"math"
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

	timers [program.Rows][program.Cols]*time.Timer
)

const (
	tickInterval = 20 * time.Millisecond
	defaultBPM   = 120
	minBPM       = 30
	maxBPM       = 480
	pulseDelay   = -1.6
	pulseRadius  = 30.0

	// if 33% of beats are active, yield to the next program
	beatLimit = program.Rows * program.Cols / 3

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

func New(ctx context.Context) program.Program {
	log := log.WithField("program", "beats")
	log.Debug("New")

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
			ledsState := leds.State{}

			// for each row, clear its full physical range:
			for _, row := range flatRows {
				for _, pixel := range row.pixels {
					ledsState.Set(pixel.strip, pixel.pixel, leds.Black)
				}
			}

			// use beatAcc to determine peak location
			for _, row := range flatRows {
				peak := float64(beatIndex) + beatAcc + pulseDelay
				if peak < 0 {
					peak += float64(program.Cols)
				}
				pulse := getPulse(row, peak)

				for coord, brightness := range pulse {
					c := leds.Brightness(leds.Mint, brightness)
					ledsState.Set(coord.strip, coord.pixel, c)
				}
			}

			// set active beats to red
			for rowIdx, beats := range beatState {
				for i, beat := range beats {
					if beat {
						redPos := flatRows[rowIdx].buttons[i]
						redIndex := flatRows[rowIdx].pixels[redPos]
						ledsState.Set(redIndex.strip, redIndex.pixel, leds.Red)
					}
				}
			}
			b.render <- ledsState

			beatAcc += beatsPerTick
			for beatAcc >= 1.0 {
				beatAcc -= 1.0

				// play active beats if index matches
				for rowIdx, beats := range beatState {
					for i, beat := range beats {
						if beat {
							if i == beatIndex {
								b.play <- sounds[rowIdx]
							}
						}
					}
				}

				// advance to next beat column
				beatIndex = (beatIndex + 1) % program.Cols
			}

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

			if press.Row < 0 || press.Row >= program.Rows || press.Col < 0 || press.Col >= program.Cols {
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
			}

			ledsState := leds.State{}
			phys := rows[press.Row].buttons[press.Col]
			ledsState.Set(phys.strip, phys.pixel, color)
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
	return program.Cols / 4.0 * float64(bpm) / 60.0 * tickInterval.Seconds()
}

// getPulse returns map of coord -> brightness
// 0 <= peak < 16
// TODO: cache results?
func getPulse(r flatRow, peak float64) map[coord]float64 {
	pulse := make(map[coord]float64)

	floatPeakPixel := peakToFloatPixel(r, peak)
	radius := pulseRadius

	startIndex := int(math.Ceil(floatPeakPixel - radius))
	endIndex := int(math.Floor(floatPeakPixel + radius))
	for i := startIndex; i <= endIndex; i++ {
		// calculate distance from peak
		distance := math.Abs(float64(i) - floatPeakPixel)
		// distance == radius => 0 brightness
		// distance == 0 => 1 brightness
		frac := 1 - distance/radius
		if frac < 0 {
			log.Errorf("FRAC <= 0 frac: %+v, distance %+v, radius %+v, getPulse(%+v, %f)", frac, distance, radius, r, peak)
			continue
		}
		// steeper falloff: raise to the 8th power for an even sharper dropoff
		bness := math.Pow(frac, 8)

		// map to physical LED index
		pixelIndex := (i + len(r.pixels)) % len(r.pixels)

		coord := coord{
			strip: r.pixels[pixelIndex].strip,
			pixel: r.pixels[pixelIndex].pixel,
		}

		pulse[coord] = bness
	}

	return pulse
}

// peakToFloatPixel converts a peak [0-16) to a float pixel value [0-143].
// 0 <= peak < 16
func peakToFloatPixel(r flatRow, peak float64) float64 {
	// assume peak == 12.7
	beat1 := math.Floor(peak) // 12 // 15
	beat2 := math.Ceil(peak)  // 13 // 16
	if beat1 == program.Cols {
		beat1 = 0.0
	}
	if beat2 == program.Cols {
		beat2 = 0.0
	}

	pixelIndex1 := r.buttons[int(beat1)] // 12 => 134
	pixelIndex2 := r.buttons[int(beat2)] // 13 => 142

	percentAhead := peak - beat1 // 12.7 - 12 => 0.7
	distance := pixelIndex2 - pixelIndex1
	if distance < 0 {
		distance = (len(r.pixels) - pixelIndex1) + pixelIndex2
	}
	diff := percentAhead * float64(distance) // 0.7 * (142 - 134) => 0.7 * 8 => 5.6

	return float64(pixelIndex1) + diff // 134 + 5.6 => 139.6
}
