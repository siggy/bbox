package leds

import (
	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

type (
	LEDs interface {
		Close() error
		Clear() error
		Write(State) error
	}

	leds struct {
		port         serial.Port
		stripLengths []int
		log          *log.Entry
	}
)

const (
	macDevicePath = "/dev/tty.usbmodem103" // macbook
	piDevicePath  = "/dev/ttyACM1"         // pi

	baudRate = 115200
)

func New(stripLengths []int, macDevice bool) (LEDs, error) {
	devicePath := piDevicePath
	if macDevice {
		devicePath = macDevicePath
	}

	log := log.WithField("leds", devicePath)

	port, err := serial.Open(devicePath, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, err
	}

	log.Infof("Connected to %s", devicePath)

	return &leds{
		port:         port,
		stripLengths: stripLengths,
		log:          log,
	}, nil
}

func (l *leds) Close() error {
	if l.port != nil {
		return l.port.Close()
	}
	return nil
}

func (l *leds) Clear() error {
	return l.write(all(l.stripLengths))
}

func (l *leds) Write(state State) error {
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

	return l.write(s)
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

func all(stripLengths []int) State {
	state := State{}
	for strip, length := range stripLengths {
		state[strip] = make(map[int]Color)
		for pixel := 0; pixel < length; pixel++ {
			state[strip][pixel] = Color{0, 0, 0, 0}
		}
	}
	return state
}
