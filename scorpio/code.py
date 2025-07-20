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
strip_lengths = [144, 144, 144, 144, 144, 144, 144, 144]
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
    # try to grab a start byte; if no data, bail immediately
    first = serial.read(1)
    if not first:
        return None
    # if it isn’t our marker, keep scanning
    while first != bytes([START_BYTE]):
        first = serial.read(1)
        if not first:
            return None
    # now read the 2-byte length
    hdr = serial.read(2)
    if not hdr or len(hdr) < 2:
        return None
    length = (hdr[0] << 8) | hdr[1]
    # read payload & checksum (non-blocking)
    payload = serial.read(length)
    if not payload or len(payload) < length:
        return None
    chk = serial.read(1)
    if not chk:
        return None
    # verify
    calc = 0
    for bb in payload:
        calc ^= bb
    if chk[0] != calc:
        print("Bad checksum")
        return None
    return payload

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
        pkt = read_packet()
        if pkt and len(pkt) % 6 == 0:
            for i in range(0, len(pkt), 6):
                si = pkt[i]
                pi = pkt[i+1]
                r, g, b, w = pkt[i+2:i+6]
                if si < TOTAL_STRIPS and pi < strip_lengths[si]:
                    strips[si][pi] = (r, g, b, w)
        # always refresh LEDs immediately
        for strip in strips:
            strip.show()
except Exception as e:
    status_led[0] = (255, 0, 0)
    print("CRASHED:", e)
