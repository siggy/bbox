package bbox

import (
	"fmt"
	"time"

	"github.com/siggy/bbox/beatboxer/wavs"
)

const (
	DEFAULT_BPM            = 120
	MIN_BPM                = 30
	MAX_BPM                = 480
	SOUNDS                 = 4
	BEATS                  = 16
	DEFAULT_TICKS_PER_BEAT = 10
	DEFAULT_TICKS          = BEATS * DEFAULT_TICKS_PER_BEAT

	TEMPO_DECAY = 3 * time.Minute
)

type Interval struct {
	TicksPerBeat int
	Ticks        int
}

type Beats [SOUNDS][BEATS]bool

type Loop struct {
	beats   Beats
	closing chan struct{}
	msgs    <-chan Beats

	bpmCh chan int
	bpm   int

	tempo      <-chan int
	tempoDecay *time.Timer

	ticks []chan<- int
	wavs  *wavs.Wavs

	iv         Interval
	intervalCh []chan<- Interval
}

var sounds = []string{
	"hihat-808.wav",
	"kick-classic.wav",
	"perc-808.wav",
	"tom-808.wav",
}

func InitLoop(
	msgs <-chan Beats,
	tempo <-chan int,
	ticks []chan<- int,
	intervalCh []chan<- Interval,
) *Loop {
	return &Loop{
		beats: Beats{},

		bpmCh: make(chan int),
		bpm:   DEFAULT_BPM,

		closing: make(chan struct{}),
		msgs:    msgs,
		tempo:   tempo,
		ticks:   ticks,
		wavs:    wavs.InitWavs(),

		intervalCh: intervalCh,
		iv: Interval{
			TicksPerBeat: DEFAULT_TICKS_PER_BEAT,
			Ticks:        DEFAULT_TICKS,
		},
	}
}

func (l *Loop) Run() {
	ticker := time.NewTicker(l.bpmToInterval(l.bpm))
	defer ticker.Stop()

	tick := 0
	tickTime := time.Now()
	for {
		select {
		case _, more := <-l.closing:
			if !more {
				fmt.Printf("Loop trying to close\n")
				// return
			}
		case beats, more := <-l.msgs:
			if more {
				// incoming beat update from keyboard
				l.beats = beats
			} else {
				// closing
				l.wavs.Close()
				fmt.Printf("Loop closing\n")
				return
			}

		case bpm, more := <-l.bpmCh:
			if more {
				// incoming bpm update
				l.bpm = bpm

				// BPM: 30 -> 60 -> 120 -> 240 -> 480.0
				// TPB: 40 -> 20 ->  10 ->   5 ->   2.5
				l.iv.TicksPerBeat = 1200 / l.bpm
				l.iv.Ticks = BEATS * l.iv.TicksPerBeat

				for _, ch := range l.intervalCh {
					ch <- l.iv
				}

				ticker.Stop()
				ticker = time.NewTicker(l.bpmToInterval(l.bpm))
				defer ticker.Stop()
			} else {
				// we should never get here
				fmt.Printf("closed on bpm, invalid state")
				panic(1)
			}

		case tempo, more := <-l.tempo:
			if more {
				// incoming tempo update from keyboard
				if (l.bpm > MIN_BPM || tempo > 0) &&
					(l.bpm < MAX_BPM || tempo < 0) {

					go l.setBpm(l.bpm + tempo)

					// set a decay timer
					if l.tempoDecay != nil {
						l.tempoDecay.Stop()
					}
					l.tempoDecay = time.AfterFunc(TEMPO_DECAY, func() {
						l.setBpm(DEFAULT_BPM)
					})
				}
			} else {
				// we should never get here
				fmt.Printf("unexpected: tempo return no more\n")
				return
			}

		case <-ticker.C: // for every time interval
			// next interval
			tick = (tick + 1) % l.iv.Ticks
			tmp := tick

			for _, ch := range l.ticks {
				ch <- tmp
			}

			// for each beat type
			if tick%l.iv.TicksPerBeat == 0 {
				for i, beat := range l.beats {
					if beat[tick/l.iv.TicksPerBeat] {
						// initiate playback
						l.wavs.Play(sounds[i])
					}
				}
			}

			t := time.Now()
			tbprint(0, 5, fmt.Sprintf("______BPM:__%+v______", l.bpm))
			tbprint(0, 6, fmt.Sprintf("______int:__%+v______", l.bpmToInterval(l.bpm)))
			tbprint(0, 7, fmt.Sprintf("______time:_%+v______", t.Sub(tickTime)))
			tbprint(0, 8, fmt.Sprintf("______tick:_%+v______", tick))
			tickTime = t
		}
	}
}

func (l *Loop) Close() {
	// TODO: this doesn't block?
	close(l.closing)
}

func (l *Loop) bpmToInterval(bpm int) time.Duration {
	return 60 * time.Second / time.Duration(bpm) / (BEATS / 4) / time.Duration(l.iv.TicksPerBeat) // 4 beats per interval
}

func (l *Loop) setBpm(bpm int) {
	l.bpmCh <- bpm
}
