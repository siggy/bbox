package wavs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ebitengine/oto/v3"
	"github.com/siggy/bbox/bbox2/equalizer"
	log "github.com/sirupsen/logrus"
	"github.com/youpy/go-wav"
)

// Alias for the new DisplayData struct from the equalizer package.
type DisplayData = equalizer.DisplayData

type Wavs struct {
	ctx     *oto.Context
	buffers map[string][]byte

	playersLock sync.Mutex
	players     []*oto.Player

	eq  *equalizer.Equalizer
	log *log.Entry
}

// EQ now correctly returns a channel of the new DisplayData struct.
func (w *Wavs) EQ() <-chan DisplayData {
	return w.eq.Data()
}

func New(dir string) (*Wavs, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Oto context: %v", err)
	}
	<-ready

	var filenames []string
	dirs, _ := os.ReadDir(dir)
	for _, d := range dirs {
		if d.IsDir() || !d.Type().IsRegular() || !strings.HasSuffix(d.Name(), ".wav") {
			continue
		}
		filenames = append(filenames, d.Name())
	}

	buffers := make(map[string][]byte)
	for _, filename := range filenames {
		filepath := filepath.Join(dir, filename)
		buf, err := fileToAudioBytes(filepath)
		if err != nil {
			return nil, fmt.Errorf("fileToAudioBytes failed: %w", err)
		}
		buffers[filename] = buf
	}

	return &Wavs{
		ctx:     ctx,
		buffers: buffers,
		eq:      equalizer.New(16),
		log:     log.WithField("bbox2", "wavs"),
	}, nil
}

func (w *Wavs) Play(filename string) {
	w.log.Tracef("play: %s", filename)
	buf, ok := w.buffers[filename]
	if !ok {
		w.log.Warnf("Unknown: %s", filename)
		return
	}

	// The equalizer now implements io.Writer, so it can be passed directly.
	reader := io.TeeReader(bytes.NewReader(buf), w.eq)
	player := w.ctx.NewPlayer(reader)
	player.Play()

	w.playersLock.Lock()
	w.players = append(w.players, player)
	w.playersLock.Unlock()
}

func (w *Wavs) StopAll() {
	w.playersLock.Lock()
	defer w.playersLock.Unlock()

	for _, player := range w.players {
		player.Close()
	}
}

func (w *Wavs) Close() {
	w.StopAll()
	w.ctx.Suspend()
	w.eq.Close()
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

	// Allow 44.1kHz or 16kHz mono audio
	if format.NumChannels != 1 || (format.SampleRate != 44100 && format.SampleRate != 16000) {
		return nil, fmt.Errorf("unsupported format (must be 44.1/16kHz mono): %v", format)
	}

	var pcm []byte
	for {
		samples, err := reader.ReadSamples()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("ReadSamples failed on file %s with error: %w", filename, err)
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
