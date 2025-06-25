package wavs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ebitengine/oto/v3"
	log "github.com/sirupsen/logrus"
	"github.com/youpy/go-wav"
)

type Wavs struct {
	ctx     *oto.Context
	buffers map[string][]byte
	players []*oto.Player

	log *log.Entry
}

func New() (*Wavs, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Oto context: %v", err)
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

	return &Wavs{
		ctx:     ctx,
		buffers: buffers,
		log:     log.WithField("bbox2", "wavs"),
	}, nil
}

func (w *Wavs) Play(filename string) {
	w.log.Debugf("Play %s", filename)
	buf, ok := w.buffers[filename]
	if !ok {
		w.log.Warnf("Unknown: %s", filename)
		return
	}

	player := getPlayer(w.ctx, buf)
	player.Play()

	// not thread safe
	w.players = append(w.players, player)
}

func (w *Wavs) StopAll() {
	for _, player := range w.players {
		player.Close()
	}
}

func (w *Wavs) Close() {
	w.StopAll()
	w.ctx.Suspend()
}

func getPlayer(ctx *oto.Context, pcm []byte) *oto.Player {
	return ctx.NewPlayer(
		bytes.NewReader(pcm),
	)
}

func fileToAudioBytes(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open failed: %w", err)
	}
	defer file.Close()

	reader := wav.NewReader(file)
	format, err := reader.Format()
	if err != nil {
		return nil, fmt.Errorf("could not get format: %w", err)
	}

	if format.NumChannels != 1 || format.SampleRate != 44100 {
		return nil, fmt.Errorf("unsupported format: %v", format)
	}

	var pcm []byte
	for {
		samples, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("ReadSamples failed on file %s with error: %s", filename, err)
			}
		}
		for _, sample := range samples {
			var sampleValue int16
			switch format.AudioFormat {
			case wav.AudioFormatPCM:
				sampleValue = int16(reader.IntValue(sample, 0))
			case wav.AudioFormatIEEEFloat:
				floatVal := reader.FloatValue(sample, 0)
				if floatVal > 1.0 {
					floatVal = 1.0
				} else if floatVal < -1.0 {
					floatVal = -1.0
				}
				sampleValue = int16(floatVal * 32767)
			default:
				return nil, fmt.Errorf("unsupported audio format code: %d", format.AudioFormat)
			}

			b := make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(sampleValue))
			pcm = append(pcm, b...)
		}
	}

	return pcm, nil
}
