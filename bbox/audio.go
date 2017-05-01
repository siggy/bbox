package bbox

import (
	"fmt"
	// "io/ioutil"
	"os"

	"github.com/gordonklaus/portaudio"
	"github.com/youpy/go-wav"
)

// type Audio struct {
// 	wavs [][]float32
// }

// func Init() *Audio {
// 	files, _ := ioutil.ReadDir("./wav")
// 	for _, f := range files {
// 		fmt.Println(f.Name())
// 	}
// 	return
// }

var out []float32

func streamCallback(out2 [][]float32) {
	fmt.Printf("streamCallback out2[0]: %+v\n", len(out2[0]))
	copy(out2[0], out)
}

func RunAudio() {
	out = make([]float32, 524288)

	file, _ := os.Open("wav/mid_tom.wav")
	reader := wav.NewReader(file)
	defer file.Close()

	samples, err := reader.ReadSamples(524288)
	for i, sample := range samples {
		out[i] = float32(reader.FloatValue(sample, 0))
	}

	portaudio.Initialize()
	// defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(samples), streamCallback)
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
