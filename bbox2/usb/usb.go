package usb

import (
	"os"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
	"golang.org/x/exp/slices"
)

const (
	devicePath = "/dev/tty.usbmodem103" // macbook
	// devicePath = "/dev/ttyACM0" // pi

	baudRate = 115200
)

var (
	// should match scorpio/code.py
	stripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}
	totalStrips  = len(stripLengths)
)

func Run() {
	port, err := serial.Open(devicePath, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		log.Errorf("Failed to open serial port: %v", err)
		os.Exit(1)
	}
	defer port.Close()

	log.Debug("Connected to device.")

	for {
		for light := 0; light < slices.Max(stripLengths); light++ {
			var payload []byte

			for strip := 0; strip < totalStrips; strip++ {
				if light >= stripLengths[strip] {
					continue
				}

				for i := 0; i < stripLengths[strip]; i++ {
					pixel := byte(i)
					g := byte(0)
					r := byte(0)
					b := byte(0)
					w := byte(0)
					if i == light {
						switch strip % 4 {
						case 0:
							g = byte(10)
						case 1:
							r = byte(10)
						case 2:
							b = byte(10)
						case 3:
							w = byte(10)
						}
					}
					payload = append(payload, byte(strip), pixel, g, r, b, w)
				}
			}

			packet := buildPacket(payload)
			n, err := port.Write(packet)
			if err != nil {
				log.Debugf("Write failed: %v", err)
				os.Exit(1)
			}

			log.Debugf("Sent %d bytes: %d pixels updated\n", n, len(payload)/6)
		}
	}
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
