package bbox

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

const (
	BPM      = 120
	BEATS    = 4
	TICKS    = 16
	INTERVAL = 60 * time.Second / BPM / (TICKS / 4) // 4 beats per interval
	BUF      = 16
	MAX_BUF  = 524288
)

type Beats [BEATS][TICKS]bool

var empty = make([]float32, BUF)

type Wav struct {
	name      string
	buf       []float32
	length    int
	stream    *portaudio.Stream
	active    chan struct{}
	remaining int
}

func InitWav(f os.FileInfo) *Wav {
	file, _ := os.Open("./wav/" + f.Name())
	reader := wav.NewReader(file)
	defer file.Close()

	buf := [MAX_BUF]float32{}
	loc := 0
	for {
		samples, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		for _, sample := range samples {
			buf[loc] = float32(reader.FloatValue(sample, 0))
			loc += 1
		}
	}

	w := Wav{
		name:   f.Name(),
		buf:    make([]float32, loc),
		length: loc,
		active: make(chan struct{}),
	}
	copy(w.buf, buf[:])

	fmt.Printf("%+v: %+v samples\n", w.name, loc)

	var err error
	w.stream, err = portaudio.OpenDefaultStream(0, 1, 44100, BUF, w.cb)
	if err != nil {
		panic(err)
	}
	err = w.stream.Start()
	if err != nil {
		panic(err)
	}

	return &w
}

func (w *Wav) Close() {
	w.stream.Stop()
	w.stream.Close()
}

func (w *Wav) cb(output [][]float32) {
	select {
	case <-w.active:
		w.remaining = w.length
	default:
	}

	if w.remaining > 0 {
		if w.remaining < BUF {
			output[0] = output[0][:w.remaining]
		}
		copy(output[0], w.buf[(w.length-w.remaining):])
		w.remaining -= BUF
	} else {
		copy(output[0], empty)
	}
}

type Audio struct {
	beats Beats
	msgs  <-chan Beats
	wavs  [BEATS]*Wav
}

func InitAudio(msgs <-chan Beats, files []os.FileInfo) *Audio {
	portaudio.Initialize()

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
					go func(j int) { a.wavs[j].active <- struct{}{} }(i)
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
