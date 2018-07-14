package wavs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	log "github.com/sirupsen/logrus"
	"github.com/youpy/go-wav"
)

const (
	WAVS = "./wavs"
	BUF  = 128
)

type Wavs struct {
	wavs   map[string]*wavFile
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
	err := portaudio.Initialize()
	if err != nil {
		panic(err)
	}

	files, _ := ioutil.ReadDir(WAVS)

	wavs := &Wavs{
		wavs: map[string]*wavFile{},
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".wav") {
			wavs.wavs[f.Name()] = initWav(f)
		}
	}

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

	buf := make([]float32, f.Size())
	loc := 0
	for {
		samples, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(fmt.Sprintf("ReadSamples failed on file %s with error: %s", f.Name(), err))
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
				// TODO: do this correctly
				// add dynamic range compressor
				out[i] += wv.buf[wv.length-wv.remaining+i]
			}
		}
		wv.remaining -= BUF
	}
	copy(output[0], out)
}

func (wavs *Wavs) Play(name string) time.Duration {
	wav, ok := wavs.wavs[name]
	if !ok {
		log.Debugf("Unknown wav file: %s", name)
		return time.Duration(0)
	}

	wav.active <- struct{}{}

	return time.Duration(wav.length*1000/44100) * time.Millisecond
}

func (w *Wavs) Close() {
	w.stream.Stop()
	w.stream.Close()
	portaudio.Terminate()

	log.Debugf("Wavs Closed")
}
