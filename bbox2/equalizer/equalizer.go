package equalizer

import (
	"bytes"
	"encoding/binary"

	// "io"
	"math"
	"sync"
	"time"

	"github.com/mjibson/go-dsp/fft"
	log "github.com/sirupsen/logrus"
)

// DisplayData  holds a history of the last four spectrum readings.
type DisplayData struct {
	History [4][]float64
}

type Equalizer struct {
	bands    int
	interval time.Duration
	data     chan DisplayData
	mu       sync.Mutex
	buf      []float64
	history  [4][]float64 // Stores the last 4 smoothed spectrums
	ticker   *time.Ticker
	quit     chan struct{}
}

const sampleRate = 44100
const fftSize = 1024

// New creates an Equalizer. The 'bands' argument is used for display resolution.
func New(bands int, interval time.Duration) *Equalizer {
	eq := &Equalizer{
		bands:    bands,
		interval: interval,
		data:     make(chan DisplayData, 1),
		quit:     make(chan struct{}),
	}
	// Initialize history slices
	for i := 0; i < 4; i++ {
		eq.history[i] = make([]float64, bands)
	}

	eq.ticker = time.NewTicker(interval)
	go eq.loop()
	return eq
}

var start = time.Now()
var addedSamples = 0
var addedSamplesTotal = 0

// AddSamples appends new PCM samples as float64s. This method is restored to fix the error.
func (eq *Equalizer) AddSamples(samples []float64) {
	addedSamples++
	addedSamplesTotal += len(samples)
	log.Infof("AddSamples called: %d samples, total calls: %d, total samples: %d, time: %v, rate: %.2f/sec", len(samples), addedSamples, addedSamplesTotal, time.Since(start), float64(addedSamplesTotal)/time.Since(start).Seconds())

	eq.mu.Lock()
	defer eq.mu.Unlock()
	eq.buf = append(eq.buf, samples...)
}

// Write implements the io.Writer interface to process raw audio bytes.
func (eq *Equalizer) Write(p []byte) (n int, err error) {
	samples := make([]float64, len(p)/2)
	reader := bytes.NewReader(p)

	for i := 0; i < len(samples); i++ {
		var sampleInt16 int16
		if err := binary.Read(reader, binary.LittleEndian, &sampleInt16); err != nil {
			// Add any samples successfully read before the error occurred
			if i > 0 {
				eq.AddSamples(samples[:i])
			}
			return i * 2, err
		}
		samples[i] = float64(sampleInt16) / 32768.0
	}

	eq.AddSamples(samples)
	return len(p), nil
}

// Data returns a channel for DisplayData.
func (eq *Equalizer) Data() <-chan DisplayData {
	return eq.data
}

// Close stops the background goroutine.
func (eq *Equalizer) Close() {
	close(eq.quit)
	eq.ticker.Stop()
	close(eq.data)
}

const hopSize = fftSize / 2 // 50% overlap

func (eq *Equalizer) loop() {
	const smoothingFactor = 0.6 // Smoothing factor for the newest spectrum.
	var smoothedSpectrum = make([]float64, eq.bands)

	ticks := 0
	start := time.Now()

	// ticksTop := 0

	for {
		select {
		case <-eq.ticker.C:
			eq.mu.Lock()

			// ticksTop++
			// log.Infof("  len(eq.buf): %d, ticks TOP: %d, time: %v, rate: %.2f/sec                     ", len(eq.buf), ticksTop, time.Since(start), float64(ticksTop)/time.Since(start).Seconds())

			// if len(eq.buf) < fftSize {
			// 	eq.mu.Unlock()
			// 	continue
			// }
			buf := eq.buf
			eq.buf = nil // we'll put leftovers back after processing
			eq.mu.Unlock()

			// window := make([]float64, fftSize)
			// copy(window, eq.buf[len(eq.buf)-fftSize:])
			// eq.buf = nil // Clear buffer after processing
			// eq.mu.Unlock()

			offset := 0
			for offset+fftSize <= len(buf) {
				window := buf[offset : offset+fftSize]
				// (optional but recommended) apply a Hann/Hamming window here before FFT
				// process(window) -> your FFT + banding + smoothing...
				offset += hopSize

				// --- Raw Signal Analysis ---
				spec := fft.FFTReal(window)
				mags := make([]float64, len(spec)/2)
				for i := 0; i < len(spec)/2; i++ {
					mags[i] = math.Hypot(real(spec[i]), imag(spec[i]))
				}

				// --- Calculate, Smooth, and Normalize Spectrum ---
				newSpectrum := calculateLogSpectrum(mags, eq.bands)
				for i := 0; i < eq.bands; i++ {
					smoothedSpectrum[i] = (newSpectrum[i] * smoothingFactor) + (smoothedSpectrum[i] * (1.0 - smoothingFactor))
				}
				normalizedFrame := normalizeSpectrum(smoothedSpectrum)

				// --- Update History ---
				// Shift old frames down
				eq.history[0] = eq.history[1]
				eq.history[1] = eq.history[2]
				eq.history[2] = eq.history[3]
				// Add the newest frame at the end
				eq.history[3] = normalizedFrame

				// --- Populate DisplayData ---
				displayData := DisplayData{
					History: eq.history,
				}

				ticks++
				log.Infof("ticks OUT: %d, time: %v, rate: %.2f/sec", ticks, time.Since(start), float64(ticks)/time.Since(start).Seconds())

				select {
				case eq.data <- displayData:
				default: // Don't block if the channel is full
				}
			}

			// Save leftover (not enough for a full frame yet) for next tick
			if offset < len(buf) {
				eq.mu.Lock()
				eq.buf = append(eq.buf, buf[offset:]...)
				eq.mu.Unlock()
			}

		case <-eq.quit:
			return
		}
	}
}

// Normalizes a spectrum slice relative to its own min and max dB levels.
func normalizeSpectrum(dbSpectrum []float64) []float64 {
	if len(dbSpectrum) == 0 {
		return nil
	}

	minDB, maxDB := -60.0, 0.0 // Use a fixed range for more stable visualization
	for _, v := range dbSpectrum {
		if v > maxDB {
			maxDB = v // Allow occasional peaks to expand the range dynamically
		}
	}

	normalized := make([]float64, len(dbSpectrum))
	dbRange := maxDB - minDB

	if dbRange < 1e-6 {
		return normalized // Avoid division by zero
	}

	for i, v := range dbSpectrum {
		norm := (v - minDB) / dbRange
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		normalized[i] = norm
	}
	return normalized
}

// Calculates the spectrum energy in dB for log-spaced frequency bands.
func calculateLogSpectrum(mags []float64, numBands int) []float64 {
	spectrum := make([]float64, numBands)
	maxFreq := float64(sampleRate) / 2.0
	binWidth := maxFreq / float64(len(mags))

	// Define the logarithmic frequency boundaries
	minLogFreq := math.Log10(20.0) // Start at 20 Hz
	maxLogFreq := math.Log10(maxFreq)
	logRange := maxLogFreq - minLogFreq

	for i := 0; i < numBands; i++ {
		// Determine frequency range for this band
		logStart := minLogFreq + (float64(i)*logRange)/float64(numBands)
		logEnd := minLogFreq + (float64(i+1)*logRange)/float64(numBands)
		freqStart := math.Pow(10, logStart)
		freqEnd := math.Pow(10, logEnd)

		// Determine which FFT bins fall into this frequency range
		binStart := int(freqStart / binWidth)
		binEnd := int(freqEnd / binWidth)
		if binEnd >= len(mags) {
			binEnd = len(mags) - 1
		}
		if binStart > binEnd {
			binStart = binEnd
		}

		// Sum the energy (squared magnitude) in the bins
		energy := 0.0
		for k := binStart; k <= binEnd; k++ {
			energy += mags[k] * mags[k]
		}
		if energy > 0 {
			spectrum[i] = 10 * math.Log10(energy)
		} else {
			spectrum[i] = -100 // Represents silence in dB
		}
	}
	return spectrum
}
