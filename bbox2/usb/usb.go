package usb

import (
	"os"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

const (
	devicePath = "/dev/tty.usbmodem103" // Adjust for your platform
	baudRate   = 115200
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
		for light := 0; light < 30; light++ {
			// Build pixel updates for LEDs 0–9
			var payload []byte
			for i := 0; i < 30; i++ {
				index := byte(i)
				g := byte(0)
				r := byte(0)
				b := byte(0)
				w := byte(0)
				if i == light {
					r = byte(10)
				}
				payload = append(payload, index, g, r, b, w)
			}

			packet := buildPacket(payload)
			port.Write(packet)
			log.Debug("Batch GRBW packet sent.")
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
