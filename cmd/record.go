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

	// -2147483648 to 2147483647
	amp := float64(0)
	last := float64(0)

	chk(stream.Start())
	for {
		chk(stream.Read())
		// fmt.Println(in)
		amp = math.Max(max(in), amp*0.95)
		db := 20 * math.Log10(amp/math.MaxInt32)
		// fmt.Printf("\r%s", strings.Repeat(" ", 100))
		level := int(db + 50)
		if level > 0 {
			fmt.Printf("%s\n", strings.Repeat("#", level))
		}
		if last < db {
			fmt.Printf("%s\n", db-last)
		}
		last = db

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
		panic(err)
	}
}
