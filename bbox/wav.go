package bbox

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

const (
	WAVS    = "./wavs"
	BUF     = 16
	MAX_BUF = 524288
)

type Wavs struct {
	wavs   [SOUNDS]*wavFile
	stream *portaudio.Stream
}

type wavFile struct {
	name      string
	buf       []float32
	length    int
	active    chan struct{}
	remaining int
}

func InitWavs() *Wavs {
	portaudio.Initialize()

	files, _ := ioutil.ReadDir(WAVS)
	if len(files) != SOUNDS {
		panic(0)
	}

	wavs := &Wavs{}

	for i, f := range files {
		wavs.wavs[i] = initWav(f)
	}

	var err error
	wavs.stream, err = portaudio.OpenDefaultStream(0, 1, 44100, BUF, wavs.cb)
	if err != nil {
		panic(err)
	}
	err = wavs.stream.Start()
	if err != nil {
		panic(err)
	}

	return wavs
}

func initWav(f os.FileInfo) *wavFile {
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

	w := wavFile{
		name:   f.Name(),
		buf:    make([]float32, loc),
		length: loc,
		active: make(chan struct{}),
	}
	copy(w.buf, buf[:])

	fmt.Printf("%+v: %+v samples\n", w.name, loc)

	return &w
}

func (w *Wavs) cb(output [][]float32) {
	for _, wv := range w.wavs {
		select {
		case <-wv.active:
			wv.remaining = wv.length
		default:
		}
	}

	out := make([]float32, BUF)
	for _, wv := range w.wavs {
		for i := 0; i < BUF; i++ {
			if wv.remaining > i {
				out[i] += wv.buf[wv.length-wv.remaining+i]
			}
		}
		wv.remaining -= BUF
	}
	copy(output[0], out)
}

func (wavs *Wavs) Play(i int) {
	wavs.wavs[i].active <- struct{}{}
}

func (w *Wavs) Close() {
	w.stream.Stop()
	w.stream.Close()
	portaudio.Terminate()

	fmt.Printf("Wavs closing\n")
}
