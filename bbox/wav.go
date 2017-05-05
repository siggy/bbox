package bbox

import (
	"fmt"
	"io"
	"os"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

const (
	BUF     = 16
	MAX_BUF = 524288
)

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
	portaudio.Initialize()

	file, _ := os.Open(WAVS + "/" + f.Name())
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

func (w *Wav) Play() {
	w.active <- struct{}{}
}

func (w *Wav) Close() {
	w.stream.Stop()
	w.stream.Close()
	portaudio.Terminate()

	fmt.Printf("%+v closed\n", w.name)
}
