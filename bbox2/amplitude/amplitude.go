package amplitude

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gen2brain/malgo"
	log "github.com/sirupsen/logrus"
)

const channelBuffer = 100

type Amplitude struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
	levels chan float64

	log *log.Entry
}

func (a *Amplitude) Level() <-chan float64 {
	return a.levels
}

func (a *Amplitude) onRecvFrames(_, in []byte, _ uint32) {
	if len(in) < 2 {
		return
	}
	var sum float64
	// S16 mono: 2 bytes per sample, little-endian
	for i := 0; i+1 < len(in); i += 2 {
		s := int16(binary.LittleEndian.Uint16(in[i : i+2]))
		// map to [-1,1]; using 32768.0 avoids overflow on -32768
		f := float64(s) / 32768.0
		sum += math.Abs(f)
	}
	n := float64(len(in) / 2) // number of samples
	avg := sum / n            // 0..1
	a.log.Tracef("Amplitude: %.3f", avg)

	select {
	case a.levels <- avg:
	default:
		a.log.Warn("Amplitude channel buffer full, dropping value")
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

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 44100

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: a.onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return nil, fmt.Errorf("failed to open capture device: %w", err)
	}
	a.device = device

	if err = a.device.Start(); err != nil {
		return nil, fmt.Errorf("failed to start capture device: %w", err)
	}

	return a, nil
}

func (a *Amplitude) Close() {
	a.log.Info("Closing")

	a.device.Uninit()
	a.ctx.Uninit()
	a.ctx.Free()
}
