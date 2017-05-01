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

type Audio struct {
	wavs    [][]float32
	lengths []int
}

func Init() *Audio {
	a := Audio{}

	files, _ := ioutil.ReadDir("./wav")
	a.wavs = make([][]float32, len(files))
	a.lengths = make([]int, len(files))

	for i, f := range files {
		a.wavs[i] = make([]float32, 524288)

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
				a.wavs[i][loc] = float32(reader.FloatValue(sample, 0))
				loc += 1
			}
		}

		fmt.Printf("%+v: %+v samples\n", f.Name(), loc)
		a.lengths[i] = loc
	}

	portaudio.Initialize()

	return &a
}

func (a *Audio) Play(i int) error {
	if i < 0 || i >= len(a.wavs) {
		return errors.New(fmt.Sprintf("index out of range: %+v", i))
	}

	var stream *portaudio.Stream
	cb := func(out2 [][]float32) {
		fmt.Printf("cb a.wavs[%+v] out2[0]: %+v\n", i, len(out2[0]))
		copy(out2[0], a.wavs[i])
		// defer stream.Close()
		// defer stream.Stop()
	}

	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, a.lengths[i], cb)
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
