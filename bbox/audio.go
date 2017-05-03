package bbox

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

const (
	BPM      = 120
	TICKS    = 16
	INTERVAL = 60 * time.Second / BPM / (TICKS / 4) // 4 beats per interval
	BUF      = 16
	MAX_BUF  = 524288
)

var empty = make([]float32, BUF)

type Wav struct {
	buf       []float32
	length    int
	stream    *portaudio.Stream
	active    chan bool
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
	wavs []*Wav
	bs   *BeatState
}

func InitAudio(bs *BeatState, files []os.FileInfo) *Audio {
	var err error

	portaudio.Initialize()

	a := Audio{
		bs:   bs,
		wavs: make([]*Wav, len(files)),
	}

	for i, f := range files {
		buf := make([]float32, MAX_BUF)

		file, _ := os.Open("./wav/" + f.Name())
		reader := wav.NewReader(file)
		defer file.Close()

		loc := 0
		for {
			samples, err := reader.ReadSamples()
			if err != io.EOF {
				chk(err)
			} else {
				break
			}

			for _, sample := range samples {
				buf[loc] = float32(reader.FloatValue(sample, 0))
				loc += 1
			}
		}

		a.wavs[i] = &Wav{
			buf:    make([]float32, loc),
			length: loc,
			active: make(chan bool),
		}
		copy(a.wavs[i].buf, buf)

		fmt.Printf("%+v: %+v samples\n", f.Name(), loc)

		a.wavs[i].stream, err = portaudio.OpenDefaultStream(0, 1, 44100, BUF, a.wavs[i].cb)
		chk(err)
		chk(a.wavs[i].stream.Start())
	}

	return &a
}

func (a *Audio) Run() {
	curTick := 0

	ticker := time.NewTicker(INTERVAL)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			for i := 0; i < len(a.wavs); i++ {
				if a.bs.Enabled(i, curTick) {
					a.Play(i)
				}
			}
			curTick = (curTick + 1) % TICKS

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (a *Audio) Play(i int) error {
	if i < 0 || i >= len(a.wavs) {
		return errors.New(fmt.Sprintf("index out of range: %+v", i))
	}
	a.wavs[i].active <- true
	return nil
}

func chk(err error) {
	if err != nil {
		fmt.Printf("died with: %+v\n", err)
		panic(err)
	}
}
