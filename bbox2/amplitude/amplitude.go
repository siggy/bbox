package amplitude

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

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

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 48000

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: a.onRecvFrames,
	}

	infos, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to get capture devices: %w", err)
	}
	for _, inf := range infos {
		if n := inf.Name(); strings.Contains(n, "USB Audio Device") {
			a.log.Infof("found usb audio device: %s", inf.String())
			deviceConfig.Capture.DeviceID = inf.ID.Pointer()
			break
		}
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

	a.device.Stop()
	a.device.Uninit()

	close(a.levels)

	a.ctx.Uninit()
	a.ctx.Free()
}
