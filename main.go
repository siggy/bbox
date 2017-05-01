package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	// "os/signal"
	"math"
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

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

var out []float32
var loc int

// buf := make([]byte, 524288)

func streamCallback(out2 [][]float32) {
	fmt.Printf("streamCallback[0]: %+v\n", len(out2[0]))
	// fmt.Printf("streamCallback[1]: %+v\n", len(out2[1]))
	// fmt.Printf("streamCallback out[0]: %+v\n", out[0])
	// fmt.Printf("streamCallback out[1]: %+v\n", out[1])
	// fmt.Printf("streamCallback out[2]: %+v\n", out[2])
	// fmt.Printf("streamCallback out[3]: %+v\n", out[3])
	// return
	// fmt.Printf("loc: %+v\n", loc)
	for i := range out2[0] {
		out2[0][i] = out[i]
		// if loc < 213093/4 {
		// 	f := Float32frombytes(out[loc*4 : loc*4+4])
		// 	out2[0][i] = f
		// 	fmt.Printf("streamCallback out[%+v : %+v]: %+v\n", loc*4, loc*4+4, out[loc*4:loc*4+4])
		// 	fmt.Printf("streamCallback f: %+v\n", f)
		// 	loc += 1
		// }
	}
}

// func streamCallback(
// 	in portaudio.Buffer,
// 	out portaudio.Buffer,
// 	timeInfo portaudio.StreamCallbackTimeInfo,
// 	flags portaudio.StreamCallbackFlags,
// ) {
// 	fmt.Printf("streamCallback\n")
// }

func runAudio() {
	out = make([]float32, 524288)
	loc = 0

	file, _ := os.Open("wav/mid_tom.wav")
	reader := wav.NewReader(file)
	defer file.Close()

	for {
		samples, err := reader.ReadSamples()
		// if err != nil {
		// 	fmt.Printf("reader.ReadSamples() errored: %+v\n", err)
		// }
		// fmt.Printf("reader.ReadSamples() returned %+v samples\n", len(samples))
		if err == io.EOF {
			fmt.Printf("reader.ReadSamples() returned EOF after %+v samples\n", loc)
			break
		}
		for _, sample := range samples {
			out[loc] = float32(reader.FloatValue(sample, 0))
			loc += 1
			// fmt.Printf("L/R: %d/%d\n", reader.IntValue(sample, 0), reader.IntValue(sample, 1))
		}
	}

	// for _, sample := range samples {
	//   t += reader.IntValue(sample, 0)
	//   t += reader.IntValue(sample, 1)
	// }

	// samples, err := reader.ReadSamples()
	// if err != nil {
	// 	fmt.Printf("reader.ReadSamples() errored: %+v\n", err)
	// }
	// fmt.Printf("reader.ReadSamples() returned %+v samples\n", len(samples))

	// for i, sample := range samples {
	// 	out[i] = float32(reader.FloatValue(sample, 0))
	// 	// fmt.Printf("L/R: %d/%d\n", reader.IntValue(sample, 0), reader.IntValue(sample, 1))
	// }

	// n, err := reader.Read(out)
	// if err != nil {
	// 	fmt.Printf("reader.Read(out) errored: %+v\n", err)
	// }
	// fmt.Printf("reader.Read(out) returned: %+v\n", n)
	// return

	// out2 := make([]byte, 262144)
	// for i, s := range out {
	// 	if i%2 != 0 {
	// 		out2[i/2] = s
	// 	}
	// }

	portaudio.Initialize()
	// defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, loc, streamCallback)
	chk(err)
	// defer stream.Close()

	chk(stream.Start())
	// defer stream.Stop()

	// chk(stream.Write())
	// for remaining := int(c.NumSamples); remaining > 0; remaining -= len(out) {
	// 	if len(out) > remaining {
	// 		out = out[:remaining]
	// 	}
	// 	err := binary.Read(audio, binary.BigEndian, out)
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	chk(err)
	// 	chk(stream.Write())
	// 	select {
	// 	case <-sig:
	// 		return
	// 	default:
	// 	}
	// }

	return

	// if n == 0 && err == io.EOF {
	// 	return
	// }

	for i, s := range out {
		fmt.Printf("out[%+v]: %+v\n", i, s)
		if i > 100 {
			break
		}
	}

	// fmt.Printf("reader.Read(out) returned: %+v\n", n)

	// samples, err := reader.ReadSamples()
	// if err == io.EOF {
	// 	break
	// }

	// for i, sample := range samples {
	// 	// fmt.Printf("L/R: %d/%d\n", reader.IntValue(sample, 0), reader.IntValue(sample, 1))
	// 	out[i] = int32(reader.IntValue(sample, 0))
	// }
	fmt.Printf("foo1")
	err = stream.Write()
	fmt.Printf("foo2")
	if err != nil {
		fmt.Printf("stream.Write() failed: %+v\n", err)
	}

	// fmt.Printf("len(samples): %+v\n", len(samples))

	return

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt, os.Kill)

	// fileName := "wav/bass_drum.wav"
	// f, err := os.Open(fileName)
	// chk(err)
	// defer f.Close()

	// id, data, err := readChunk(f)
	// chk(err)
	// if id.String() != "FORM" {
	// 	fmt.Println("bad file format")
	// 	return
	// }
	// _, err = data.Read(id[:])
	// chk(err)
	// if id.String() != "AIFF" {
	// 	fmt.Println("bad file format")
	// 	return
	// }
	// var c commonChunk
	// var audio io.Reader
	// for {
	// 	id, chunk, err := readChunk(data)
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	chk(err)
	// 	switch id.String() {
	// 	case "COMM":
	// 		chk(binary.Read(chunk, binary.BigEndian, &c))
	// 	case "SSND":
	// 		chunk.Seek(8, 1) //ignore offset and block
	// 		audio = chunk
	// 	default:
	// 		fmt.Printf("ignoring unknown chunk '%s'\n", id)
	// 	}
	// }

	// //assume 44100 sample rate, mono, 32 bit

	// portaudio.Initialize()
	// defer portaudio.Terminate()
	// out := make([]int32, 8192)
	// stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(out), &out)
	// chk(err)
	// defer stream.Close()

	// chk(stream.Start())
	// defer stream.Stop()
	// for remaining := int(c.NumSamples); remaining > 0; remaining -= len(out) {
	// 	if len(out) > remaining {
	// 		out = out[:remaining]
	// 	}
	// 	err := binary.Read(audio, binary.BigEndian, out)
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	chk(err)
	// 	chk(stream.Write())
	// 	select {
	// 	case <-sig:
	// 		return
	// 	default:
	// 	}
	// }
}

func readChunk(r readerAtSeeker) (id ID, data *io.SectionReader, err error) {
	_, err = r.Read(id[:])
	if err != nil {
		return
	}
	var n int32
	err = binary.Read(r, binary.BigEndian, &n)
	fmt.Println(err)
	fmt.Println(n)
	if err != nil {
		return
	}
	off, _ := r.Seek(0, 1)
	data = io.NewSectionReader(r, off, int64(n))
	_, err = r.Seek(int64(n), 1)
	fmt.Println(err)
	fmt.Println(n)
	return
}

type readerAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type ID [4]byte

func (id ID) String() string {
	return string(id[:])
}

type commonChunk struct {
	NumChans      int16
	NumSamples    int32
	BitsPerSample int16
	SampleRate    [10]byte
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// go runInput()
	runAudio()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
