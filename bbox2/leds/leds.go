package leds

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

type (
	Color struct {
		R uint8 // 0-255
		G uint8 // 0-255
		B uint8 // 0-255
		W uint8 // 0-255, white channel for RGBW LEDs
	}

	// map[strip][pixel]Color
	leds map[int]map[int]Color
)

const (
	devicePath = "/dev/tty.usbmodem103" // macbook
	// devicePath = "/dev/ttyACM1" // pi

	baudRate = 115200
)

type Strips struct {
	port         serial.Port
	stripLengths []int
	ledBuffer    leds
}

func New(stripLengths []int) (*Strips, error) {
	port, err := serial.Open(devicePath, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, err
	}
	log.Debugf("Connected to %s", devicePath)
	return &Strips{
		port:         port,
		stripLengths: stripLengths,
		ledBuffer:    leds{},
	}, nil
}

func (s *Strips) Close() error {
	if s.port != nil {
		return s.port.Close()
	}
	return nil
}

func (s *Strips) Clear() {
	if err := s.write(s.all()); err != nil {
		log.Errorf("Failed to clear LEDs: %v", err)
	}
}

func (s *Strips) Set(strip int, pixel int, color Color) error {
	if strip < 0 || strip >= len(s.stripLengths) {
		return fmt.Errorf("Invalid strip index: %d", strip)
	}
	if pixel < 0 || pixel >= s.stripLengths[strip] {
		return fmt.Errorf("Invalid pixel index: %d for strip %d", pixel, strip)
	}

	if _, ok := s.ledBuffer[strip]; !ok {
		s.ledBuffer[strip] = make(map[int]Color)
	}
	s.ledBuffer[strip][pixel] = color

	return nil
}

func (s *Strips) Write() error {
	err := s.write(s.ledBuffer)
	s.ledBuffer = leds{}
	return err
}

func (s *Strips) all() leds {
	leds := leds{}
	for strip, length := range s.stripLengths {
		leds[strip] = make(map[int]Color)
		for pixel := 0; pixel < length; pixel++ {
			leds[strip][pixel] = Color{0, 0, 0, 0}
		}
	}
	return leds
}

func (s *Strips) write(leds leds) error {
	var payload []byte

	for strip, stripLEDs := range leds {
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
