package bbox

import (
	"fmt"
	"math"
	"sync"

	"github.com/gordonklaus/portaudio"
)

type Amplitude struct {
	level chan<- float64
	wg    *sync.WaitGroup
}

const (
	SMOOTHING = 0.99
)

var (
	volMax        = 0.001
	stereoVol     = float64(0)
	MAX_SMOOTHING = math.Pow(SMOOTHING, 1.0/100)
)

// taken from:
// https://github.com/processing/p5.js/blob/master/lib/addons/p5.sound.js#L2305
func amp(slice []int32) float64 {
	bufLength := float64(len(slice))

	sum := float64(0)
	for _, n := range slice {
		x := math.Abs(float64(n) / math.MaxInt32)
		sum += math.Pow(math.Max(math.Min(float64(x)/volMax, 1), -1), 2)
	}
	rms := math.Sqrt(sum / bufLength)
	stereoVol = math.Max(rms, stereoVol*SMOOTHING)
	volMax = math.Max(stereoVol, volMax*MAX_SMOOTHING)
	stereoVolNorm := math.Max(math.Min(stereoVol/volMax, 1), 0)
	return stereoVolNorm
}

func InitAmplitude(wg *sync.WaitGroup, level chan<- float64) *Amplitude {
	wg.Add(1)

	return &Amplitude{
		level: level,
		wg:    wg,
	}
}

func (a *Amplitude) Run() {
	defer a.wg.Done()

	err := portaudio.Initialize()
	if err != nil {
		fmt.Printf("portaudio.Initialize failed: %+v", err)
		panic(err)
	}
	defer portaudio.Terminate()

	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	if err != nil {
		fmt.Printf("OpenDefaultStream failed: %+v", err)
		panic(err)
	}
	defer stream.Close()

	err = stream.Start()
	if err != nil {
		fmt.Printf("stream.Start failed: %+v", err)
		panic(err)
	}
	defer stream.Stop()

	for {
		// this returns `Input overflowed` sometimes, ignore it
		stream.Read()
		a.level <- amp(in)
	}
}
