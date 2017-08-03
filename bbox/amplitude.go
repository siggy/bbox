package bbox

import (
	"fmt"
	"math"
	"time"

	"github.com/gordonklaus/portaudio"
)

type Amplitude struct {
	closing chan struct{}
	level   chan<- float64
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

func InitAmplitude(level chan<- float64) *Amplitude {
	return &Amplitude{
		closing: make(chan struct{}),
		level:   level,
	}
}

func (a *Amplitude) Run() {
	err := portaudio.Initialize()
	if err != nil {
		fmt.Printf("portaudio.Initialize failed: %+v\n", err)
		panic(err)
	}
	defer portaudio.Terminate()

	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	if err != nil {
		fmt.Printf("OpenDefaultStream failed: %+v\n", err)
		panic(err)
	}
	defer stream.Close()

	err = stream.Start()
	if err != nil {
		fmt.Printf("stream.Start failed: %+v\n", err)
		panic(err)
	}
	defer stream.Stop()

	for {
		// this returns `Input overflowed` sometimes, ignore it
		stream.Read()
		select {
		case _, more := <-a.closing:
			if !more {
				return
			}
		case a.level <- amp(in):
		case <-time.After(1 * time.Millisecond):
			// default:
			// fmt.Printf("Amplitude: No message send\n")
		}
	}
}

func (a *Amplitude) Close() {
	close(a.closing)
}
