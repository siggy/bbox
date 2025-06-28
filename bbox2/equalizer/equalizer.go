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

// assume 44.1 kHz PCM
const sampleRate = 44100

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
	const fftSize = 1024
	for {
		select {
		case <-eq.ticker.C:
			// grab a snapshot, **then** remove only the slice we process
			eq.mu.Lock()
			buf := eq.buf
			if len(buf) > fftSize {
				// drop only the first fftSize samples
				eq.buf = buf[fftSize:]
				buf = buf[:fftSize]
			} else {
				// not enough for a full window next time
				eq.buf = nil
			}
			eq.mu.Unlock()

			// We only FFT on a fixed, power-of-two window (say 1024 samples).
			const fftSize = 1024
			if len(buf) < fftSize {
				// not enough data yet
				continue
			}
			// take the most recent fftSize samples
			window := buf[len(buf)-fftSize:]

			// FFTReal on a power-of-two length is much faster
			spec := fft.FFTReal(window)
			half := fftSize / 2

			// compute magnitudes
			mags := make([]float64, half)
			for i := 0; i < half; i++ {
				re := real(spec[i])
				im := imag(spec[i])
				mags[i] = math.Hypot(re, im)
			}

			// bucket into eq.bands linear-frequency bands, convert to dB, then normalize
			// define evenly spaced frequency edges from 0 up to Nyquist (sampleRate/2)
			fmax := float64(sampleRate) / 2
			edgesHz := make([]float64, eq.bands+1)
			for i := 0; i <= eq.bands; i++ {
				edgesHz[i] = fmax * float64(i) / float64(eq.bands)
			}
			// map to FFT bin indices
			binFreq := float64(sampleRate) / float64(fftSize)
			edges := make([]int, len(edgesHz))
			for i, hz := range edgesHz {
				edges[i] = int(hz/binFreq + 0.5)
			}
			bandData := make([]float64, eq.bands)
			for b := 0; b < eq.bands; b++ {
				start, end := edges[b], edges[b+1]
				if start < 0 {
					start = 0
				}
				if end > half {
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
				// average magnitude
				avg := sum / float64(width)
				// convert to decibels
				bandData[b] = 20 * math.Log10(avg+1e-8)
			}
			// normalize 0..1
			minDB, maxDB := bandData[0], bandData[0]
			for _, v := range bandData {
				if v < minDB {
					minDB = v
				}
				if v > maxDB {
					maxDB = v
				}
			}
			rangeDB := maxDB - minDB
			if rangeDB <= 0 {
				rangeDB = 1
			}
			for i, v := range bandData {
				bandData[i] = (v - minDB) / rangeDB
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
