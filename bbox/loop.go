package bbox

import (
	"io/ioutil"
	"time"
)

const (
	BPM      = 120
	BEATS    = 4
	TICKS    = 16
	INTERVAL = 60 * time.Second / BPM / (TICKS / 4) // 4 beats per interval
	WAVS     = "./wavs"
)

type Beats [BEATS][TICKS]bool

type Loop struct {
	beats Beats
	msgs  <-chan Beats
	wavs  [BEATS]*Wav
	ticks chan<- int
}

func InitLoop(msgs <-chan Beats, ticks chan<- int) *Loop {
	l := Loop{
		beats: Beats{},
		msgs:  msgs,
		wavs:  [BEATS]*Wav{},
		ticks: ticks,
	}

	files, _ := ioutil.ReadDir(WAVS)
	if len(files) != BEATS {
		panic(0)
	}

	for i, f := range files {
		l.wavs[i] = InitWav(f)
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
					l.wavs[i].Play()
				}
			}

			// next interval
			tick = (tick + 1) % TICKS
			tmp := tick
			l.ticks <- tmp
		}
	}
}

func (l *Loop) Close() {
	for _, wav := range l.wavs {
		wav.Close()
	}
}
