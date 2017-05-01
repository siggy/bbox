package main

import (
	// "encoding/binary"
	"fmt"
	// "io"
	"os"
	// "os/signal"
	// "math"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/nsf/termbox-go"
	"github.com/youpy/go-wav"
)

func runInput() {
	var current string
	var curev termbox.Event

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt)

	data := make([]byte, 0, 64)
mainloop:
	for {
		if cap(data)-len(data) < 32 {
			newdata := make([]byte, len(data), len(data)+32)
			copy(newdata, data)
			data = newdata
		}
		beg := len(data)
		d := data[beg : beg+32]
		switch ev := termbox.PollRawEvent(d); ev.Type {
		case termbox.EventRaw:
			data = data[:beg+ev.N]
			current = fmt.Sprintf("%q", data)
			if current == `"q"` {
				panic(0)
				break mainloop
			}

			fmt.Println(data)
			fmt.Println(current)
			fmt.Println(curev)

			for {
				ev := termbox.ParseEvent(data)
				fmt.Printf("  data: %+v\n", data)
				fmt.Printf("  ev: %+v\n", ev)

				if ev.N == 0 {
					break
				}
				curev = ev
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

var out []float32
var loc int

func streamCallback(out2 [][]float32) {
	fmt.Printf("streamCallback out2[0]: %+v\n", len(out2[0]))
	copy(out2[0], out)
}

func runAudio() {
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

func main() {
	// go runInput()
	go runAudio()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
