package equalizer

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mjibson/go-dsp/fft"
)

// DisplayData holds the final normalized values for all four bars.
type DisplayData struct {
	Kurtosis      float64
	Spectrum      []float64 // This will now hold frame-normalized values (0.0 to 1.0)
	LowFreqEnergy float64
	DenoisedLevel float64
}

type Equalizer struct {
	bands            int
	interval         time.Duration
	data             chan DisplayData
	mu               sync.Mutex
	buf              []float64
	smoothedSpectrum []float64 // This will store smoothed dB values
	ticker           *time.Ticker
	quit             chan struct{}
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	stdout           io.ReadCloser
}

const sampleRate = 44100
const fftSize = 1024
const activityThreshold = 0.05

// New creates an Equalizer. The 'bands' argument is now used for display resolution.
func New(bands int, interval time.Duration) *Equalizer {
	cmd := exec.Command("python3", "-u", "speech_enhancer.py")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start speech_enhancer.py: %v", err)
	}
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				log.Printf("[python-stderr]: %s", buf[:n])
			}
		}
	}()
	eq := &Equalizer{
		bands:            bands,
		interval:         interval,
		data:             make(chan DisplayData, 1),
		quit:             make(chan struct{}),
		smoothedSpectrum: make([]float64, bands),
		cmd:              cmd,
		stdin:            stdin,
		stdout:           stdout,
	}
	eq.ticker = time.NewTicker(interval)
	go eq.loop()
	return eq
}

// AddSamples appends new PCM samples.
func (eq *Equalizer) AddSamples(samples []float64) {
	eq.mu.Lock()
	eq.buf = append(eq.buf, samples...)
	eq.mu.Unlock()
}

// Data returns a channel for DisplayData.
func (eq *Equalizer) Data() <-chan DisplayData {
	return eq.data
}

// Close stops the background goroutine and the Python process.
func (eq *Equalizer) Close() {
	close(eq.quit)
	eq.ticker.Stop()
	eq.stdin.Close()
	eq.cmd.Wait()
	close(eq.data)
}

func (eq *Equalizer) loop() {
	const smoothingFactor = 0.5 // Higher is faster
	for {
		select {
		case <-eq.ticker.C:
			eq.mu.Lock()
			if len(eq.buf) < fftSize {
				eq.mu.Unlock()
				continue
			}
			window := make([]float64, fftSize)
			copy(window, eq.buf[len(eq.buf)-fftSize:])
			eq.buf = nil
			eq.mu.Unlock()

			// --- Denoised Signal ---
			enhancedWindow, err := eq.enhanceSpeech(window)
			if err != nil {
				if err != io.EOF && !strings.Contains(err.Error(), "broken pipe") {
					log.Printf("Error enhancing speech: %v", err)
				}
				continue
			}
			denoisedLevel := calculateRMS(enhancedWindow) / activityThreshold
			if denoisedLevel > 1.0 {
				denoisedLevel = 1.0
			}

			

			// --- Raw Signal Analysis ---
			spec := fft.FFTReal(window)
			mags := make([]float64, len(spec)/2)
			for i := 0; i < len(spec)/2; i++ {
				mags[i] = math.Hypot(real(spec[i]), imag(spec[i]))
			}

			// --- Calculate & Smooth Spectrum (in dB) ---
			newSpectrum := calculateLogSpectrum(mags, eq.bands)
			for i := 0; i < eq.bands; i++ {
				eq.smoothedSpectrum[i] = (newSpectrum[i] * smoothingFactor) + (eq.smoothedSpectrum[i] * (1.0 - smoothingFactor))
			}

			// --- Populate DisplayData ---
			displayData := DisplayData{
				Kurtosis:      normalizeKurtosis(calculateKurtosis(window)),
				Spectrum:      normalizeSpectrum(eq.smoothedSpectrum), // Apply per-frame normalization here
				LowFreqEnergy: normalizeEnergy(calculateFrequencyEnergy(mags, 200)),
				DenoisedLevel: denoisedLevel,
			}

			select {
			case eq.data <- displayData:
			default:
			}
		case <-eq.quit:
			return
		}
	}
}

// enhanceSpeech sends audio to the Python process.
func (eq *Equalizer) enhanceSpeech(samples []float64) ([]float64, error) {
	buf := new(bytes.Buffer)
	for _, s := range samples {
		sampleInt16 := int16(s * 32767)
		binary.Write(buf, binary.LittleEndian, sampleInt16)
	}
	payload := buf.Bytes()
	lenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBytes, uint32(len(payload)))
	if _, err := eq.stdin.Write(lenBytes); err != nil {
		return nil, err
	}
	if _, err := eq.stdin.Write(payload); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(eq.stdout, lenBytes); err != nil {
		return nil, err
	}
	respLen := binary.LittleEndian.Uint32(lenBytes)
	respPayload := make([]byte, respLen)
	if _, err := io.ReadFull(eq.stdout, respPayload); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(respPayload)
	numSamples := int(respLen) / 2
	processedSamples := make([]float64, numSamples)
	for i := range processedSamples {
		var sampleInt16 int16
		binary.Read(reader, binary.LittleEndian, &sampleInt16)
		processedSamples[i] = float64(sampleInt16) / 32767.0
	}
	return processedSamples, nil
}

// Normalizes a spectrum slice relative to its own min and max.
func normalizeSpectrum(dbSpectrum []float64) []float64 {
	if len(dbSpectrum) == 0 {
		return nil
	}

	minDB := dbSpectrum[0]
	maxDB := dbSpectrum[0]
	for _, v := range dbSpectrum {
		if v < minDB {
			minDB = v
		}
		if v > maxDB {
			maxDB = v
		}
	}

	normalized := make([]float64, len(dbSpectrum))
	dbRange := maxDB - minDB

	if dbRange < 1e-6 {
		return normalized
	}

	for i, v := range dbSpectrum {
		normalized[i] = (v - minDB) / dbRange
	}
	return normalized
}

// Calculates the spectrum energy in dB for log-spaced frequency bands.
func calculateLogSpectrum(mags []float64, numBands int) []float64 {
	spectrum := make([]float64, numBands)
	maxFreq := float64(sampleRate) / 2.0
	binWidth := maxFreq / float64(len(mags))
	minLogFreq := math.Log10(20.0)
	maxLogFreq := math.Log10(maxFreq)
	logRange := maxLogFreq - minLogFreq

	for i := 0; i < numBands; i++ {
		logStart := minLogFreq + (float64(i) * logRange / float64(numBands))
		logEnd := minLogFreq + (float64(i+1) * logRange / float64(numBands))
		freqStart := math.Pow(10, logStart)
		freqEnd := math.Pow(10, logEnd)
		binStart := int(freqStart / binWidth)
		binEnd := int(freqEnd / binWidth)
		if binEnd >= len(mags) {
			binEnd = len(mags) - 1
		}
		if binStart > binEnd {
			binStart = binEnd
		}
		energy := 0.0
		for k := binStart; k <= binEnd; k++ {
			energy += mags[k] * mags[k]
		}
		if energy > 0 {
			spectrum[i] = 10 * math.Log10(energy)
		} else {
			spectrum[i] = -100
		}
	}
	return spectrum
}

// Calculates the energy below a given frequency in Hz.
func calculateFrequencyEnergy(mags []float64, freqHz float64) float64 {
	binWidth := float64(sampleRate) / float64(fftSize)
	targetBin := int(freqHz / binWidth)
	if targetBin >= len(mags) {
		targetBin = len(mags)
	}
	energy := 0.0
	for i := 1; i < targetBin; i++ {
		energy += mags[i] * mags[i]
	}
	return energy
}

// Calculates the kurtosis ("peakiness") of the audio signal.
func calculateKurtosis(samples []float64) float64 {
	n := float64(len(samples))
	if n == 0 {
		return 0.0
	}
	mean := 0.0
	for _, s := range samples {
		mean += s
	}
	mean /= n
	m2, m4 := 0.0, 0.0
	for _, s := range samples {
		dev := s - mean
		m2 += dev * dev
		m4 += dev * dev * dev * dev
	}
	m2 /= n
	m4 /= n
	if m2 == 0 {
		return 0.0
	}
	return m4 / (m2 * m2)
}

// Calculates the Root Mean Square (average power) of the signal.
func calculateRMS(samples []float64) float64 {
	if len(samples) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, s := range samples {
		sum += s * s
	}
	return math.Sqrt(sum / float64(len(samples)))
}

// Normalizes energy based on a fixed maximum.
func normalizeEnergy(e float64) float64 {
	const maxEnergy = 250.0
	norm := e / maxEnergy
	if norm > 1.0 {
		return 1.0
	}
	return norm
}

// Normalizes kurtosis to the 0.0-1.0 range.
func normalizeKurtosis(k float64) float64 {
	const minK, maxK = 1.0, 8.0
	norm := (k - minK) / (maxK - minK)
	if norm < 0 {
		return 0.0
	}
	if norm > 1 {
		return 1.0
	}
	return norm
}