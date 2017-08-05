package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
)

func max(slice []int32) float64 {
	result := float64(0)
	for _, n := range slice {
		abs := math.Abs(float64(n))
		if abs > result {
			result = abs
		}
	}
	return result
}

func rms(slice []int32) float64 {
	sum := float64(0)
	for _, n := range slice {
		sum += float64(n) * float64(n)
	}
	return math.Sqrt(sum / float64(len(slice)))
}

func mean(slice []int32) float64 {
	sum := float64(0)
	for _, n := range slice {
		sum += math.Abs(float64(n))
	}

	return sum / float64(len(slice))
}

func median(slice []int32) float64 {
	numbers := make([]float64, len(slice))
	for i, n := range slice {
		numbers[i] = math.Abs(float64(n))
	}
	sort.Float64s(numbers)

	middle := len(numbers) / 2
	result := numbers[middle]
	if len(numbers)%2 == 0 {
		result = (result + numbers[middle-1]) / 2
	}
	return result
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

func main() {
	fmt.Println("Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	for {
		chk(stream.Read())
		level := amp(in)
		fmt.Printf("\r%s", strings.Repeat(" ", 100))
		fmt.Printf("\r%s", strings.Repeat("#", int(100*level)))

		select {
		case <-sig:
			return
		default:
		}
	}
	chk(stream.Stop())
}

func chk(err error) {
	if err != nil {
		// panic(err)
	}
}
