# SPDX-FileCopyrightText: 2021 Kattni Rembor for Adafruit Industries
#
# SPDX-License-Identifier: MIT

import board
import neopixel
import usb_cdc

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

while True:
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
