package bbox

import (
	"time"
)

const (
	BPM      = 120
	BEATS    = 4
	TICKS    = 16
	INTERVAL = 60 * time.Second / BPM / (TICKS / 4) // 4 beats per interval
)

type Beats [BEATS][TICKS]bool

type Loop struct {
	beats Beats
	msgs  <-chan Beats
	wavs  *Wavs
	ticks []chan<- int
}

func InitLoop(msgs <-chan Beats, ticks []chan<- int) *Loop {
	l := Loop{
		beats: Beats{},
		msgs:  msgs,
		wavs:  InitWavs(),
		ticks: ticks,
	}

	return &l
}

func (l *Loop) Run() {
	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()
	defer l.Close()

	tick := 0
	for {
		select {
		case beats, more := <-l.msgs:
			if more {
				// incoming beat update from keyboard
				l.beats = beats
			} else {
				// closing
				return
			}
		case <-ticker.C: // for every time interval
			// for each beat type
			for i, beat := range l.beats {
				if beat[tick] {
					// initiate playback
					l.wavs.Play(i)
				}
			}

			// next interval
			tick = (tick + 1) % TICKS
			tmp := tick

			for _, ch := range l.ticks {
				ch <- tmp
			}
		}
	}
}

func (l *Loop) Close() {
	l.wavs.Close()
}
