# SPDX-FileCopyrightText: 2021 Kattni Rembor for Adafruit Industries
#
# SPDX-License-Identifier: MIT

import board
import neopixel
import usb_cdc

NUM_STRIPS = 1
PIXELS_PER_STRIP = 30
TOTAL_PIXELS = NUM_STRIPS * PIXELS_PER_STRIP

pixels = neopixel.NeoPixel(
    board.NEOPIXEL0,
    PIXELS_PER_STRIP,
    pixel_order=neopixel.GRBW,
    bpp=4,
    auto_write=False
)
pixels.fill((0, 0, 0, 0))
pixels.show()

serial = usb_cdc.data  # Use secondary CDC interface for data
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
        len_bytes = read_exact(2)
        length = (len_bytes[0] << 8) | len_bytes[1]
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
    # print("Waiting for packet")
    packet = read_packet()
    # print("Got packet: ", packet)
    if packet and len(packet) % 5 == 0:
        for i in range(0, len(packet), 5):
            led_index = packet[i]
            g, r, b, w = packet[i+1:i+5]
            if led_index < TOTAL_PIXELS:
                pixels[led_index] = (r, g, b, w)
        pixels.show()
