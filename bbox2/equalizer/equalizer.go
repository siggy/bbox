package equalizer

import (
	"math"
	"sync"
	"time"

	"github.com/mjibson/go-dsp/fft"
)

type Equalizer struct {
	bands    int
	interval time.Duration
	data     chan []float64

	mu  sync.Mutex
	buf []float64

	ticker *time.Ticker
	quit   chan struct{}
}

// New creates an Equalizer that divides the spectrum into `bands`
// and emits one slice of band‐amplitudes every `interval`.
func New(bands int, interval time.Duration) *Equalizer {
	eq := &Equalizer{
		bands:    bands,
		interval: interval,
		data:     make(chan []float64, 1),
		quit:     make(chan struct{}),
	}
	eq.ticker = time.NewTicker(interval)
	go eq.loop()
	return eq
}

// AddSamples appends new PCM samples (as floats in range [-1,1]).
// You should call this from your audio playback callback or wherever
// you have access to the decoded audio frames.
func (eq *Equalizer) AddSamples(samples []float64) {
	eq.mu.Lock()
	eq.buf = append(eq.buf, samples...)
	eq.mu.Unlock()
}

// Data returns a channel that produces one []float64 per interval,
// each of length `bands`, giving the average magnitude in that band.
func (eq *Equalizer) Data() <-chan []float64 {
	return eq.data
}

// Close stops the background goroutine and closes the Data channel.
func (eq *Equalizer) Close() {
	close(eq.quit)
	eq.ticker.Stop()
	close(eq.data)
}

func (eq *Equalizer) loop() {
	for {
		select {
		case <-eq.ticker.C:
			eq.mu.Lock()
			buf := eq.buf
			eq.buf = nil
			eq.mu.Unlock()

			// Skip if we don't have enough samples for a decent FFT
			const minFFTsize = 512
			if len(buf) < minFFTsize {
				continue
			}

			// FFT
			spec := fft.FFTReal(buf)
			half := len(spec) / 2

			// compute magnitudes
			mags := make([]float64, half)
			for i := 0; i < half; i++ {
				re := real(spec[i])
				im := imag(spec[i])
				mags[i] = math.Hypot(re, im)
			}

			// bucket into bands
			bandData := make([]float64, eq.bands)
			step := half / eq.bands
			for b := 0; b < eq.bands; b++ {
				start := b * step
				end := start + step
				if b == eq.bands-1 {
					end = half
				}
				width := end - start
				if width <= 0 {
					bandData[b] = 0
					continue
				}
				sum := 0.0
				for i := start; i < end; i++ {
					sum += mags[i]
				}
				bandData[b] = sum / float64(width)
			}

			// emit (non-blocking if consumer is slow)
			select {
			case eq.data <- bandData:
			default:
			}

		case <-eq.quit:
			return
		}
	}
}
