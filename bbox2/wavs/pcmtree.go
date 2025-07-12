package wavs

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/siggy/bbox/bbox2/equalizer"
)

// pcmTeeReader wraps an underlying PCM byte reader, and feeds decoded
// float samples into an equalizer before returning the raw bytes.
type pcmTeeReader struct {
	r   io.Reader
	eq  *equalizer.Equalizer
	buf []byte
}

func newPCMTeeReader(pcm []byte, eq *equalizer.Equalizer) *pcmTeeReader {
	return &pcmTeeReader{
		r:   bytes.NewReader(pcm),
		eq:  eq,
		buf: make([]byte, 4096),
	}
}

func (t *pcmTeeReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	if n > 0 {
		// convert each 2-byte LE sample into a float and feed the EQ
		samples := make([]float64, n/2)
		for i := 0; i+1 < n; i += 2 {
			v := int16(binary.LittleEndian.Uint16(p[i : i+2]))
			samples[i/2] = float64(v) / 32767.0
		}
		t.eq.AddSamples(samples)
	}
	return n, err
}
