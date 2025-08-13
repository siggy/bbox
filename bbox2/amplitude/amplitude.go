package amplitude

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
	log "github.com/sirupsen/logrus"
)

const channelBuffer = 100

type Amplitude struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
	levels chan float64

	deviceConfig    malgo.DeviceConfig    // saved so we can reinit
	deviceCallbacks malgo.DeviceCallbacks // saved so we can reinit
	lastNanos       int64                 // last time we saw audio
	mu              sync.Mutex            // guards device (stop/uninit/reinit)
	log             *log.Entry
}

func (a *Amplitude) Level() <-chan float64 {
	return a.levels
}

func (a *Amplitude) onRecvFrames(_, in []byte, _ uint32) {
	if len(in) < 2 {
		return
	}
	atomic.StoreInt64(&a.lastNanos, time.Now().UnixNano())

	var sum float64
	// S16 mono LE: 2 bytes per sample
	for i := 0; i < len(in)-1; i += 2 {
		s := int16(binary.LittleEndian.Uint16(in[i:]))
		f := float64(s) / 32768.0 // [-1,1]
		sum += math.Abs(f)
	}

	n := float64(len(in) / 2) // number of samples
	avg := sum / n            // 0..1

	s := ""
	for range int(avg * 100) {
		s += "█"
	}
	a.log.Tracef("%.3f: %s\n", avg, s)

	select {
	case a.levels <- avg:
	default:
		a.log.Trace("Amplitude channel buffer full, dropping value")
	}
}

func New() (*Amplitude, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create malgo context: %w", err)
	}

	a := &Amplitude{
		ctx:    ctx,
		levels: make(chan float64, channelBuffer),
		log:    log.WithField("bbox2", "amplitude"),
	}

	a.deviceConfig = malgo.DefaultDeviceConfig(malgo.Capture)
	a.deviceConfig.Capture.Format = malgo.FormatS16
	a.deviceConfig.Capture.Channels = 1
	a.deviceConfig.SampleRate = 48000 // prefer 48k for USB codecs

	a.deviceCallbacks = malgo.DeviceCallbacks{
		Data: a.onRecvFrames,
	}

	infos, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to get capture devices: %w", err)
	}
	for _, inf := range infos {
		if n := inf.Name(); strings.Contains(n, "USB Audio Device") {
			a.log.Infof("found usb audio device: %s", inf.String())
			a.deviceConfig.Capture.DeviceID = inf.ID.Pointer()
			break
		}
	}

	device, err := malgo.InitDevice(ctx.Context, a.deviceConfig, a.deviceCallbacks)
	if err != nil {
		return nil, fmt.Errorf("failed to open capture device: %w", err)
	}
	a.device = device

	if err = a.device.Start(); err != nil {
		return nil, fmt.Errorf("failed to start capture device: %w", err)
	}

	// Watchdog: if no audio for > 10 seconds, restart the device
	atomic.StoreInt64(&a.lastNanos, time.Now().UnixNano())
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for range t.C {
			last := time.Unix(0, atomic.LoadInt64(&a.lastNanos))
			if time.Since(last) > 10*time.Second {
				a.log.Warn("no audio received for 10 seconds; restarting capture device")
				a.restartCapture()
				// After restart, give it a fresh window
				atomic.StoreInt64(&a.lastNanos, time.Now().UnixNano())
			}
		}
	}()

	return a, nil
}

func (a *Amplitude) restartCapture() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.device != nil {
		_ = a.device.Stop()
		a.device.Uninit()
	}

	dev, err := malgo.InitDevice(a.ctx.Context, a.deviceConfig, a.deviceCallbacks)
	if err != nil {
		a.log.WithError(err).Error("failed to re-open capture device")
		return
	}
	a.device = dev

	if err := a.device.Start(); err != nil {
		a.log.WithError(err).Error("failed to restart capture device")
		// best effort: uninit the half-open device
		a.device.Uninit()
		a.device = nil
		return
	}

	a.log.Info("capture device restarted")
}

func (a *Amplitude) Close() {
	a.log.Info("Closing")

	a.mu.Lock()
	if a.device != nil {
		_ = a.device.Stop()
		a.device.Uninit()
	}
	a.mu.Unlock()

	close(a.levels)

	a.ctx.Uninit()
	a.ctx.Free()
}
