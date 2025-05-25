package bbox2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type Wav struct {
	ctx     *oto.Context
	buffers map[string][]byte
	players []*oto.Player
}

func Init() (*Wav, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create Oto context: %v", err)
	}
	<-ready

	filenames := []string{}
	dirs, _ := os.ReadDir("./wavs")
	for _, d := range dirs {
		if d.IsDir() || !d.Type().IsRegular() || !strings.HasSuffix(d.Name(), ".wav") {
			continue
		}

		filenames = append(filenames, d.Name())
	}

	buffers := map[string][]byte{}
	for _, filename := range filenames {
		buf, err := fileToAudioBytes("./wavs/" + filename)
		if err != nil {
			return nil, fmt.Errorf("fileToAudioBytes failed: %s", err)
		}
		buffers[filename] = buf
	}

	return &Wav{
		ctx:     ctx,
		buffers: buffers,
	}, nil
}

func (w *Wav) Play(filename string) {
	buf, ok := w.buffers[filename]
	if !ok {
		return
	}

	player := getPlayer(w.ctx, buf)
	player.Play()

	// not thread safe
	w.players = append(w.players, player)
}

func (w *Wav) StopAll() {
	for _, player := range w.players {
		player.Close()
	}
}

func (w *Wav) Close() {
	w.StopAll()
	w.ctx.Suspend()
}

func getPlayer(ctx *oto.Context, pcm []byte) *oto.Player {
	return ctx.NewPlayer(
		bytes.NewReader(pcm),
	)
}

func fileToAudioBytes(filename string) ([]byte, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	decoder := wav.NewDecoder(bytes.NewReader(fileBytes))
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file: %s", filename)
	}

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	if buf.Format.SampleRate == 0 || buf.Format.NumChannels == 0 {
		return nil, fmt.Errorf("missing WAV format information: %v", buf.Format)
	}

	// fmt.Printf("filename: %s\n", filename)
	// fmt.Printf("Audio format code: %d\n", decoder.WavAudioFormat)
	// fmt.Printf("WAV file format: %v\n", buf.Format)

	pcm := intBufferTo16LEBytes(buf)

	return pcm, nil
}

func intBufferTo16LEBytes(buf *audio.IntBuffer) []byte {
	out := make([]byte, len(buf.Data)*2)
	for i, sample := range buf.Data {
		if sample > 32767 {
			sample = 32767
		} else if sample < -32768 {
			sample = -32768
		}
		binary.LittleEndian.PutUint16(out[i*2:], uint16(int16(sample)))
	}
	return out
}
