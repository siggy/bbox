package leds

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

const (
	devicePath = "/dev/tty.usbmodem103" // macbook
	// devicePath = "/dev/ttyACM1" // pi

	baudRate = 115200
)

type Leds struct {
	port         serial.Port
	stripLengths []int
}

func New(stripLengths []int) (*Leds, error) {
	port, err := serial.Open(devicePath, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, err
	}
	log.Debugf("Connected to %s", devicePath)
	return &Leds{
		port:         port,
		stripLengths: stripLengths,
	}, nil
}

func (s *Leds) Close() error {
	if s.port != nil {
		return s.port.Close()
	}
	return nil
}

func (s *Leds) Clear() error {
	return s.write(s.all())
}

func (s *Leds) Write(state State) error {
	for strip, stripLEDs := range state {
		if strip < 0 || strip >= len(s.stripLengths) {
			return fmt.Errorf("invalid strip index: %d", strip)
		}
		for pixel := range stripLEDs {
			if pixel < 0 || pixel >= s.stripLengths[strip] {
				return fmt.Errorf("invalid pixel index: %d for strip %d", pixel, strip)
			}
		}
	}
	return s.write(state)
}

func (s *Leds) all() State {
	state := State{}
	for strip, length := range s.stripLengths {
		state[strip] = make(map[int]Color)
		for pixel := 0; pixel < length; pixel++ {
			state[strip][pixel] = Color{0, 0, 0, 0}
		}
	}
	return state
}

func (s *Leds) write(state State) error {
	var payload []byte

	for strip, stripLEDs := range state {
		for pixel, color := range stripLEDs {
			payload = append(payload, byte(strip), byte(pixel), color.R, color.G, color.B, color.W)
		}
	}

	packet := buildPacket(payload)
	n, err := s.port.Write(packet)
	if err != nil {
		return err
	}

	log.Debugf("Sent %d bytes: %d pixels updated\n", n, len(payload)/6)
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
