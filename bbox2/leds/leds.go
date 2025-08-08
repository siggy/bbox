package leds

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

type (
	LEDs interface {
		Close() error
		Clear()
		Set(State)
	}

	leds struct {
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup

		set          chan State
		port         serial.Port
		stripLengths []int
		log          *log.Entry
	}
)

const (
	macDevicePath = "/dev/tty.usbmodem1101" // macbook
	piDevicePath  = "/dev/ttyACM1"          // pi

	baudRate = 115200

	setBuffer         = 1000
	tickInterval      = 30 * time.Millisecond
	reconcileInterval = 60 * time.Second
)

func New(ctx context.Context, stripLengths []int, macDevice bool) (LEDs, error) {
	devicePath := piDevicePath
	if macDevice {
		devicePath = macDevicePath
	}

	log := log.WithField("leds", devicePath)

	port, err := serial.Open(devicePath, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", devicePath, err)
	}

	log.Infof("Connected to %s", devicePath)

	ctx, cancel := context.WithCancel(ctx)
	l := &leds{
		ctx:    ctx,
		cancel: cancel,

		set:          make(chan State, setBuffer),
		port:         port,
		stripLengths: stripLengths,
		log:          log,
	}

	l.wg.Add(1)
	go l.run()

	return l, nil
}

func (l *leds) Close() error {
	l.cancel()
	l.wg.Wait()
	close(l.set)

	if l.port != nil {
		return l.port.Close()
	}
	return nil
}

func (l *leds) Clear() {
	l.Set(l.all())
}

func (l *leds) Set(state State) {
	select {
	case <-l.ctx.Done():
		return // driver is shutting down, ignore
	default:
	}

	s := State{}

	for strip, stripLEDs := range state {
		if strip < 0 || strip >= len(l.stripLengths) {
			l.log.Warnf("invalid strip index: %d", strip)
			continue
		}

		for pixel, color := range stripLEDs {
			if pixel < 0 || pixel >= l.stripLengths[strip] {
				l.log.Warnf("invalid pixel index: %d for strip %d", pixel, strip)
				continue
			}

			s.Set(strip, pixel, color)
		}
	}

	select {
	case l.set <- s:
	default:
		// drop the update if we’re backed up
	}
}

func (l *leds) run() {
	defer l.wg.Done()

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	reconcile := time.NewTicker(reconcileInterval)
	defer reconcile.Stop()

	currentState := l.all()
	lastTick := l.all()

	// clear all at startup
	l.write(currentState)

	ticks := 0
	last := time.Now()
	tickTime := time.Duration(0)

	for {
		select {
		case <-ticker.C:
			ticks++
			tickTime += time.Since(last)
			last = time.Now()
			if ticks%100 == 0 {
				l.log.Debugf("Tick %d, average tick time: %v", ticks, tickTime/time.Duration(ticks))
				l.log.Tracef("lastTick: %v", lastTick)
			}
			// send a diff of the LEDs
			if err := l.write(lastTick.diff(currentState)); err != nil {
				l.log.Errorf("Failed to reconcile full state: %v", err)
				continue
			}

			lastTick = currentState.copy()

		case <-reconcile.C:
			// send the full state to the LEDs
			if err := l.write(currentState); err != nil {
				l.log.Errorf("Failed to reconcile full state: %v", err)
				continue
			}

			lastTick = currentState.copy()

		case s := <-l.set:
			currentState.ApplyState(s)
		case <-l.ctx.Done():
			return
		}
	}
}

func (l *leds) write(state State) error {
	var payload []byte

	for strip, stripLEDs := range state {
		for pixel, color := range stripLEDs {
			payload = append(payload, byte(strip), byte(pixel), color.R, color.G, color.B, color.W)
		}
	}

	packet := buildPacket(payload)
	n, err := l.port.Write(packet)
	if err != nil {
		return err
	}

	l.log.Tracef("Sent %d bytes: %d pixels updated\n", n, len(payload)/6)
	return nil
}

func buildPacket(payload []byte) []byte {
	length := len(payload)
	lengthHi := byte((length >> 8) & 0xFF)
	lengthLo := byte(length & 0xFF)

	packet := []byte{0xAA, lengthHi, lengthLo}
	packet = append(packet, payload...)

	checksum := byte(0)
	for _, b := range payload {
		checksum ^= b
	}
	packet = append(packet, checksum)

	return packet
}

func (l *leds) all() State {
	state := State{}
	for strip, length := range l.stripLengths {
		state[strip] = make(map[int]Color)
		for pixel := range length {
			state[strip][pixel] = Color{0, 0, 0, 0}
		}
	}
	return state
}

func all(stripLengths []int) State {
	state := State{}
	for strip, length := range stripLengths {
		state[strip] = make(map[int]Color)
		for pixel := range length {
			state[strip][pixel] = Color{0, 0, 0, 0}
		}
	}
	return state
}
