package bbox

import (
	"fmt"
	"time"
)

const (
	BPM            = 120
	SOUNDS         = 4
	BEATS          = 16
	TICKS_PER_BEAT = 10
	TICKS          = BEATS * TICKS_PER_BEAT
	INTERVAL       = 60 * time.Second / BPM / (BEATS / 4) / TICKS_PER_BEAT // 4 beats per interval
)

type Beats [SOUNDS][BEATS]bool

type Loop struct {
	beats   Beats
	closing chan struct{}
	msgs    <-chan Beats
	tempo   <-chan int
	ticks   []chan<- int
	wavs    *Wavs
}

func InitLoop(msgs <-chan Beats, tempo <-chan int, ticks []chan<- int) *Loop {
	return &Loop{
		beats:   Beats{},
		closing: make(chan struct{}),
		msgs:    msgs,
		tempo:   tempo,
		ticks:   ticks,
		wavs:    InitWavs(),
	}
}

func (l *Loop) Run() {
	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	tick := 0
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

		case tempo, more := <-l.tempo:
			if more {
				// incoming tempo update from keyboard
				fmt.Printf("TEMPO: %+v", tempo)
			} else {
				// we should never get here
				fmt.Printf("unexpected: tempo return no more\n")
				return
			}

		case <-ticker.C: // for every time interval
			// next interval
			tick = (tick + 1) % TICKS
			tmp := tick

			for _, ch := range l.ticks {
				ch <- tmp
			}

			// for each beat type
			if tick%TICKS_PER_BEAT == 0 {
				for i, beat := range l.beats {
					if beat[tick/TICKS_PER_BEAT] {
						// initiate playback
						l.wavs.Play(i)
					}
				}
			}
		}
	}
}

func (l *Loop) Close() {
	// TODO: this doesn't block?
	close(l.closing)
}
