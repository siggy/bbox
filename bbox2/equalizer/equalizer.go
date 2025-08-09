package equalizer

import (
	"bytes"
	"encoding/binary"

	"math"
	"sync"
	"time"

	"github.com/mjibson/go-dsp/fft"
	log "github.com/sirupsen/logrus"
)

// DisplayData  holds a history of the last four spectrum readings.
type DisplayData struct {
	History [HistorySize][]float64
}

type Equalizer struct {
	bands   int
	data    chan DisplayData
	mu      sync.Mutex
	cond    *sync.Cond // <— add
	buf     []float64
	history [HistorySize][]float64 // Stores the last 4 smoothed
	quit    chan struct{}
}

const HistorySize = 4
const sampleRate = 44100
const fftSize = 1024

// New creates an Equalizer. The 'bands' argument is used for display resolution.
func New(bands int) *Equalizer {
	eq := &Equalizer{
		bands: bands,
		data:  make(chan DisplayData, 1),
		quit:  make(chan struct{}),
	}
	// Initialize history slices
	for i := range HistorySize {
		eq.history[i] = make([]float64, bands)
	}

	eq.cond = sync.NewCond(&eq.mu)
	go eq.loop()
	return eq
}

var start = time.Now()
var last = time.Now()
var addedSamples = 0
var addedSamplesTotal = 0

// AddSamples appends new PCM samples as float64s. This method is restored to fix the error.
func (eq *Equalizer) AddSamples(samples []float64) {
	addedSamples++
	addedSamplesTotal += len(samples)
	log.Infof("AddSamples called: %d samples, total calls: %d, total samples: %d, time: %v, rate: %.2f/sec", len(samples), addedSamples, addedSamplesTotal, time.Since(start), float64(addedSamplesTotal)/time.Since(start).Seconds())
	log.Info("Time since last AddSamples: ", time.Since(last))
	last = time.Now()

	eq.mu.Lock()
	eq.buf = append(eq.buf, samples...)
	eq.cond.Signal() // signal *while holding* eq.mu
	eq.mu.Unlock()
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
	eq.mu.Lock()
	eq.cond.Broadcast()
	eq.mu.Unlock()
	close(eq.data)
}

const hopSize = fftSize / 2 // 50% overlap

func (eq *Equalizer) loop() {
	const smoothingFactor = 0.6
	smoothedSpectrum := make([]float64, eq.bands)

	ticks := 0
	start := time.Now()

	// Pace by audio time: ~86.1 hops/sec for 44.1k, 1024/2.
	hopDur := time.Second * time.Duration(hopSize) / time.Duration(sampleRate)
	next := time.Now()

	for {
		// Wait until we have enough for one frame (or we're told to quit).
		eq.mu.Lock()
		for len(eq.buf) < fftSize {
			// Wake on AddSamples (Signal) or Close (Broadcast).
			eq.cond.Wait()
			select {
			case <-eq.quit:
				eq.mu.Unlock()
				return
			default:
			}
		}

		// Take the oldest fftSize samples to preserve temporal order.
		window := make([]float64, fftSize)
		copy(window, eq.buf[:fftSize])

		if hopSize > len(eq.buf) { // defensive
			eq.buf = eq.buf[:0]
		} else {
			eq.buf = eq.buf[hopSize:]
		}
		eq.mu.Unlock()

		// --- Raw Signal Analysis ---
		// (Optional) apply a Hann window to `window` here to reduce leakage.
		spec := fft.FFTReal(window)
		mags := make([]float64, len(spec)/2)
		for i := 0; i < len(mags); i++ {
			mags[i] = math.Hypot(real(spec[i]), imag(spec[i]))
		}

		// --- Calculate, Smooth, and Normalize Spectrum ---
		newSpectrum := calculateLogSpectrum(mags, eq.bands)
		for i := 0; i < eq.bands; i++ {
			smoothedSpectrum[i] = newSpectrum[i]*smoothingFactor + smoothedSpectrum[i]*(1.0-smoothingFactor)
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

		// --- Pace to audio clock & cap latency if behind ---
		next = next.Add(hopDur)
		if d := time.Until(next); d > 0 {
			// Sleep (interruptible by quit).
			timer := time.NewTimer(d)
			select {
			case <-timer.C:
			case <-eq.quit:
				timer.Stop()
				return
			}
		} else if d < -3*hopDur {
			// We're > ~3 hops behind: drop some oldest samples to avoid display lag.
			eq.mu.Lock()
			hopsBehind := int((-d) / hopDur)
			drop := hopsBehind * hopSize
			// Keep at least one full frame to avoid starving the next cycle.
			if keep := len(eq.buf) - drop; keep < fftSize {
				drop = len(eq.buf) - fftSize
			}
			if drop > 0 {
				if drop > len(eq.buf) {
					drop = len(eq.buf)
				}
				eq.buf = eq.buf[drop:]
			}
			eq.mu.Unlock()
			next = time.Now()
		}

		// Allow quit between cycles.
		select {
		case <-eq.quit:
			return
		default:
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
