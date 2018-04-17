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
	MIN_MAX_VOL = 0.1

	SMOOTHING_FAST = 0.9
	SMOOTHING_SLOW = 0.99
)

var (
	vol           = float64(0)
	volMax        = MIN_MAX_VOL
	MAX_SMOOTHING = math.Pow(0.999, 1.0/100)

	firstRun = true
)

// taken from:
// https://github.com/processing/p5.js/blob/master/lib/addons/p5.sound.js#L2305
func amp(slice []int32) float64 {
	bufLength := float64(len(slice))

	sum := float64(0)
	for _, n := range slice {
		x := math.Abs(float64(n) / math.MaxInt32)
		sum += math.Pow(math.Min(float64(x)/volMax, 1), 2)
	}
	rms := math.Sqrt(sum / bufLength)

	if firstRun && rms > 0 {
		volMax = rms
		firstRun = false
	}

	if rms > volMax {
		volMax = (1-SMOOTHING_FAST)*rms + volMax*SMOOTHING_FAST
	} else {
		volMax = (1-SMOOTHING_SLOW)*rms + volMax*SMOOTHING_SLOW
	}

	return volMax
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
		}
	}
}

func (a *Amplitude) Close() {
	close(a.closing)
}
