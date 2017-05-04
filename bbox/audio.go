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
	quit  chan struct{}
}

func InitAudio(msgs <-chan Beats, files []os.FileInfo) *Audio {
	var err error

	portaudio.Initialize()

	a := Audio{
		beats: Beats{},
		msgs:  msgs,
		wavs:  [BEATS]*Wav{},
	}

	for i, f := range files {
		buf := make([]float32, MAX_BUF)

		file, _ := os.Open("./wav/" + f.Name())
		reader := wav.NewReader(file)
		defer file.Close()

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

		a.wavs[i] = &Wav{
			name:   f.Name(),
			buf:    make([]float32, loc),
			length: loc,
			active: make(chan struct{}),
		}
		copy(a.wavs[i].buf, buf)

		fmt.Printf("%+v: %+v samples\n", a.wavs[i].name, loc)

		a.wavs[i].stream, err = portaudio.OpenDefaultStream(0, 1, 44100, BUF, a.wavs[i].cb)
		if err != nil {
			panic(err)
		}
		err = a.wavs[i].stream.Start()
		if err != nil {
			panic(err)
		}
	}

	return &a
}

func (a *Audio) Run() {
	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	cur := 0
	for {
		select {
		case beats := <-a.msgs:
			// incoming beat update from keyboard
			a.beats = beats
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
