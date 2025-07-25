import board
import math
import time
import usb_cdc
from adafruit_led_animation.helper import PixelMap
from adafruit_neopxl8 import NeoPxl8
from neopixel import NeoPixel

# ——— USB CDC Setup ———
ser = usb_cdc.data

# ——— Heartbeat LED ———
hb = NeoPixel(board.NEOPIXEL, 1, brightness=0.5, auto_write=True)
hb[0] = (0, 255, 0)

# ——— LED Strips Setup ———
num_strands = 8
strand_length = 144
first_led_pin = board.NEOPIXEL0

num_pixels = num_strands * strand_length

# Make the object to control the pixels
pixels = NeoPxl8(
    first_led_pin,
    num_pixels,
    num_strands=num_strands,
    auto_write=False,
    bpp=4,
)

def strand(n):
    return PixelMap(
        pixels,
        range(n * strand_length, (n + 1) * strand_length),
        individual_pixels=True,
    )

# Create the 8 virtual strands
strands = [strand(i) for i in range(num_strands)]

# Startup sequence: show all red for 1 second
pixels.fill((1, 0, 0, 0))
pixels.show()
time.sleep(1)

# clear after sequence
pixels.fill(0)
pixels.show()

# ——— Rolling serial buffer ———
buf = bytearray()
START = 0xAA

def update_heartbeat():
    t = time.monotonic()
    pulse = (math.sin(t * 2 * math.pi) + 1) / 2
    hb[0] = (0, int(pulse * 64), int((1 - pulse) * 64))

# ——— Main Loop ———
while True:
    update_heartbeat()

    # read any incoming bytes
    if ser and ser.in_waiting:
        buf.extend(ser.read(ser.in_waiting))

    # parse complete frames
    while True:
        if len(buf) < 4:
            break
        if buf[0] != START:
            # drop until next 0xAA
            buf[:] = buf[1:]
            continue

        length = (buf[1] << 8) | buf[2]
        total  = 3 + length + 1
        if len(buf) < total:
            break

        payload = buf[3 : 3 + length]
        chk     = buf[3 + length]

        # verify checksum
        x = 0
        for b in payload:
            x ^= b

        if x == chk and (length % 6) == 0:
            # apply pixel updates
            for i in range(0, length, 6):
                si, pi, r, g, b_, w = payload[i : i + 6]
                if si < num_strands and pi < strand_length:
                    strands[si][pi] = (r, g, b_, w)

        # drop processed bytes
        buf[:] = buf[total:]

    # render all pixels
    pixels.show()
