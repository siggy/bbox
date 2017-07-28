package main

import (
	"encoding/binary"
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
	if len(os.Args) < 2 {
		fmt.Println("missing required argument:  output file name")
		return
	}
	fmt.Println("Recording.  Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	fileName := os.Args[1]
	if !strings.HasSuffix(fileName, ".aiff") {
		fileName += ".aiff"
	}
	f, err := os.Create(fileName)
	chk(err)

	// form chunk
	_, err = f.WriteString("FORM")
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(0))) //total bytes
	_, err = f.WriteString("AIFF")
	chk(err)

	// common chunk
	_, err = f.WriteString("COMM")
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(18)))                  //size
	chk(binary.Write(f, binary.BigEndian, int16(1)))                   //channels
	chk(binary.Write(f, binary.BigEndian, int32(0)))                   //number of samples
	chk(binary.Write(f, binary.BigEndian, int16(32)))                  //bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bit sample rate 44100
	chk(err)

	// sound chunk
	_, err = f.WriteString("SSND")
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(0))) //size
	chk(binary.Write(f, binary.BigEndian, int32(0))) //offset
	chk(binary.Write(f, binary.BigEndian, int32(0))) //block
	nSamples := 0
	defer func() {
		// fill in missing sizes
		totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
		_, err = f.Seek(4, 0)
		chk(err)
		chk(binary.Write(f, binary.BigEndian, int32(totalBytes)))
		_, err = f.Seek(22, 0)
		chk(err)
		chk(binary.Write(f, binary.BigEndian, int32(nSamples)))
		_, err = f.Seek(42, 0)
		chk(err)
		chk(binary.Write(f, binary.BigEndian, int32(4*nSamples+8)))
		chk(f.Close())
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	chk(err)
	defer stream.Close()

	// -2147483648 to 2147483647
	amp := float64(0)

	chk(stream.Start())
	for {
		chk(stream.Read())
		chk(binary.Write(f, binary.BigEndian, in))
		nSamples += len(in)
		// fmt.Println(in)
		amp = math.Max(max(in), amp*0.95)
		db := 20 * math.Log10(amp/math.MaxInt32)
		fmt.Printf("\r%s", strings.Repeat(" ", 100))
		level := int(db + 50)
		if level > 0 {
			fmt.Printf("\r%s", strings.Repeat("#", level))
		}

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
