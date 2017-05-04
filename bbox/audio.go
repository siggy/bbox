package bbox

import (
	"os"
	"time"
)

const (
	BPM      = 120
	BEATS    = 4
	TICKS    = 16
	INTERVAL = 60 * time.Second / BPM / (TICKS / 4) // 4 beats per interval
)

type Beats [BEATS][TICKS]bool

type Audio struct {
	beats Beats
	msgs  <-chan Beats
	wavs  [BEATS]*Wav
}

func InitAudio(msgs <-chan Beats, files []os.FileInfo) *Audio {
	a := Audio{
		beats: Beats{},
		msgs:  msgs,
		wavs:  [BEATS]*Wav{},
	}

	for i, f := range files {
		a.wavs[i] = InitWav(f)
	}

	return &a
}

func (a *Audio) Run() {
	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()
	defer a.Close()

	cur := 0
	for {
		select {
		case beats, more := <-a.msgs:
			if more {
				// incoming beat update from keyboard
				a.beats = beats
			} else {
				// closing
				return
			}
		case <-ticker.C: // for every time interval
			// for each beat type
			for i, beat := range a.beats {
				if beat[cur] {
					// initiate playback
					a.wavs[i].Play()
				}
			}
			// next interval
			cur = (cur + 1) % TICKS
		}
	}
}

func (a *Audio) Close() {
	for _, wav := range a.wavs {
		wav.Close()
	}
}
