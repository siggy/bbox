package bbox

import (
	"sync"
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
	wg    *sync.WaitGroup
	beats Beats
	msgs  <-chan Beats
	wavs  *Wavs
	ticks []chan<- int
}

func InitLoop(wg *sync.WaitGroup, msgs <-chan Beats, ticks []chan<- int) *Loop {
	wg.Add(1)

	return &Loop{
		wg:    wg,
		beats: Beats{},
		msgs:  msgs,
		wavs:  InitWavs(),
		ticks: ticks,
	}
}

func (l *Loop) Run() {
	defer l.wg.Done()

	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	tick := 0
	for {
		select {
		case beats, more := <-l.msgs:
			if more {
				// incoming beat update from keyboard
				l.beats = beats
			} else {
				// closing
				l.wavs.Close()
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
