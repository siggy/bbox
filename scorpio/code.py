# SPDX-FileCopyrightText: 2021 Kattni Rembor for Adafruit Industries
#
# SPDX-License-Identifier: MIT

import board
import math
import neopixel
import time
import usb_cdc

# Onboard status LED
status_led = neopixel.NeoPixel(board.NEOPIXEL, 1, brightness=0.5, auto_write=True)
status_led[0] = (0, 255, 0)  # Green = healthy

# Number of pixels on each strip
strip_lengths = [30, 30, 10, 10, 10, 10, 10, 10]
TOTAL_STRIPS = len(strip_lengths)

# Map each output pin
output_pins = [
    board.NEOPIXEL0,
    board.NEOPIXEL1,
    board.NEOPIXEL2,
    board.NEOPIXEL3,
    board.NEOPIXEL4,
    board.NEOPIXEL5,
    board.NEOPIXEL6,
    board.NEOPIXEL7,
]

# Create one NeoPixel object per output strip
strips = [
    neopixel.NeoPixel(
        pin,
        strip_lengths[i],
        bpp=4,
        pixel_order=neopixel.GRBW,
        auto_write=False
    )
    for i, pin in enumerate(output_pins)
]

# Clear all strips
for strip in strips:
    strip.fill((0, 0, 0, 0))
    strip.show()

serial = usb_cdc.data
serial.timeout = 0.1

START_BYTE = 0xAA

def read_exact(count):
    data = bytearray()
    while len(data) < count:
        chunk = serial.read(count - len(data))
        if chunk:
            data.extend(chunk)
    return data

def read_packet():
    while True:
        if serial.read(1) != bytes([START_BYTE]):
            continue
        length_bytes = read_exact(2)
        length = (length_bytes[0] << 8) | length_bytes[1]
        payload = read_exact(length)
        checksum = serial.read(1)
        if not checksum:
            return None
        calc = 0
        for b in payload:
            calc ^= b
        if checksum[0] == calc:
            return payload
        else:
            print("Bad checksum")
            return None

def update_heartbeat():
    t = time.monotonic()  # seconds since power-on
    # Create a pulsing brightness using sine wave (0 to 1)
    pulse = (math.sin(t * 2 * math.pi) + 1) / 2  # range 0–1
    # Modulate green and blue channels
    g = int(pulse * 64)       # 0–64
    b = int((1 - pulse) * 64) # 64–0
    status_led[0] = (0, g, b)

try:
    while True:
        update_heartbeat()
        packet = read_packet()
        if packet and len(packet) % 6 == 0:
            for i in range(0, len(packet), 6):
                strip_index = packet[i]
                pixel_index = packet[i+1]
                g, r, b, w = packet[i+2:i+6]
                if strip_index < TOTAL_STRIPS:
                    if pixel_index < strip_lengths[strip_index]:
                        strips[strip_index][pixel_index] = (r, g, b, w)
            for strip in strips:
                strip.show()
except Exception as e:
    status_led[0] = (255, 0, 0)  # Red = error
    print("CRASHED:", e)
