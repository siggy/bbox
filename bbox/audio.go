package bbox

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

type Wav struct {
	buf    []float32
	stream *portaudio.Stream
	active bool
}

func (w *Wav) cb(output [][]float32) {
	fmt.Printf("cb: %+v\n", len(output[0]))
	if w.active {
		fmt.Printf("  w.active: %+v\n", len(output[0]))
		copy(output[0], w.buf)
		w.active = false
	} else {
		copy(output[0], make([]float32, len(output[0])))
	}
	fmt.Printf("  output[0][0] %+v\n", output[0][0])
}

type Audio struct {
	wavs []*Wav
}

func Init() *Audio {
	var err error

	a := Audio{}

	portaudio.Initialize()

	files, _ := ioutil.ReadDir("./wav")
	a.wavs = make([]*Wav, len(files))
	// a.wavs = make([][]float32, len(files))
	// a.lengths = make([]int, len(files))

	for i, f := range files {
		buf := make([]float32, 524288)

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
			buf: make([]float32, loc),
		}
		copy(a.wavs[i].buf, buf)

		fmt.Printf("%+v: %+v samples\n", f.Name(), loc)
		// a.lengths[i] = loc

		// var stream *portaudio.Stream
		// cb := func(output [][]float32) {
		// 	fmt.Printf("%+v: cb a.wavs[%+v] output[0]: %+v\n", f.Name(), i, len(output[0]))
		// 	bufLen := len(output[0])
		// 	for _, wav := range a.wavs {
		// 		if bufLen == len(wav.buf) {
		// 			if wav.active {
		// 				fmt.Printf("wav.active: %+v: cb a.wavs[%+v] output[0]: %+v\n", f.Name(), i, len(output[0]))
		// 				copy(output[0], wav.buf)
		// 				wav.active = false
		// 			}
		// 			fmt.Printf("output[0][0] %+v\n", output[0][0])
		// 			break
		// 		}
		// 	}
		// 	// copy(output[0], a.wavs[i].buf)
		// 	// defer stream.Close()
		// 	// defer stream.Stop()
		// }

		a.wavs[i].stream, err = portaudio.OpenDefaultStream(0, 1, 44100, len(a.wavs[i].buf), a.wavs[i].cb)
		chk(err)
		chk(a.wavs[i].stream.Start())
	}

	return &a
}

func (a *Audio) Play(i int) error {
	if i < 0 || i >= len(a.wavs) {
		return errors.New(fmt.Sprintf("index out of range: %+v", i))
	}

	a.wavs[i].active = true
	return nil

	var stream *portaudio.Stream
	cb := func(out2 [][]float32) {
		fmt.Printf("cb a.wavs[%+v] out2[0]: %+v\n", i, len(out2[0]))
		copy(out2[0], a.wavs[i].buf)
		// defer stream.Close()
		// defer stream.Stop()
	}

	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(a.wavs[i].buf), cb)
	chk(err)
	chk(stream.Start())

	return nil
}

var out []float32

func streamCallback(out2 [][]float32) {
	fmt.Printf("streamCallback out2[0]: %+v\n", len(out2[0]))
	copy(out2[0], out)
}

func RunAudio() {
	return

	out = make([]float32, 524288)

	file, _ := os.Open("wav/bass_drum.wav")
	reader := wav.NewReader(file)
	defer file.Close()

	loc := 0

	for {
		samples, err := reader.ReadSamples()
		if len(samples) == 0 {
			break
		}

		for _, sample := range samples {
			out[loc] = float32(reader.FloatValue(sample, 0))
			loc += 1
		}

		if err == io.EOF {
			fmt.Printf("break2\n")
			break
		}
	}
	fmt.Printf("reader.ReadSamples read %+v\n", loc)

	portaudio.Initialize()
	// defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, loc, streamCallback)
	chk(err)
	// defer stream.Close()

	chk(stream.Start())
	// defer stream.Stop()

	return
}

func chk(err error) {
	if err != nil {
		fmt.Printf("died with: %+v\n", err)
		panic(err)
	}
}
