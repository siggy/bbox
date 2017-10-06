package bbox

import (
	"fmt"
	"time"
)

const (
	DEFAULT_BPM    = 120
	MIN_BPM        = 30
	MAX_BPM        = 480
	SOUNDS         = 4
	BEATS          = 16
	TICKS_PER_BEAT = 10
	TICKS          = BEATS * TICKS_PER_BEAT
)

type Beats [SOUNDS][BEATS]bool

type Loop struct {
	beats   Beats
	bpm     int
	closing chan struct{}
	msgs    <-chan Beats
	tempo   <-chan int
	ticks   []chan<- int
	wavs    *Wavs
}

func InitLoop(msgs <-chan Beats, tempo <-chan int, ticks []chan<- int) *Loop {
	return &Loop{
		beats:   Beats{},
		bpm:     DEFAULT_BPM,
		closing: make(chan struct{}),
		msgs:    msgs,
		tempo:   tempo,
		ticks:   ticks,
		wavs:    InitWavs(),
	}
}

func (l *Loop) Run() {
	ticker := time.NewTicker(bpmToInterval(l.bpm))
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
				if (l.bpm > MIN_BPM || tempo > 0) &&
					(l.bpm < MAX_BPM || tempo < 0) {
					l.bpm += tempo
					ticker.Stop()
					ticker = time.NewTicker(bpmToInterval(l.bpm))
					defer ticker.Stop()

					fmt.Printf("BPM: %+v", l.bpm)
				}
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

func bpmToInterval(bpm int) time.Duration {
	return 60 * time.Second / time.Duration(bpm) / (BEATS / 4) / TICKS_PER_BEAT // 4 beats per interval
}
