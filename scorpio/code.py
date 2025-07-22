import board
import math
import time
import usb_cdc
import digitalio
from neopixel import NeoPixel, neopixel_write

# ——— USB CDC Setup ———
ser = usb_cdc.data

# ——— Heartbeat LED ———
hb = NeoPixel(board.NEOPIXEL, 1, brightness=0.5, auto_write=True)
hb[0] = (0, 255, 0)

# ——— LED Strips Setup ———
strip_lengths = [144] * 8
TOTAL_STRIPS  = len(strip_lengths)

output_pins = [
    board.NEOPIXEL0, board.NEOPIXEL1, board.NEOPIXEL2, board.NEOPIXEL3,
    board.NEOPIXEL4, board.NEOPIXEL5, board.NEOPIXEL6, board.NEOPIXEL7,
]

# Prepare raw buffers and digital IO pins
buffers = [bytearray(4 * L) for L in strip_lengths]
dirty   = [True] * TOTAL_STRIPS  # mark all for initial write
ios     = []
for pin in output_pins:
    dio = digitalio.DigitalInOut(pin)
    dio.direction = digitalio.Direction.OUTPUT
    ios.append(dio)

# Clear at startup
for i, dio in enumerate(ios):
    neopixel_write(dio, buffers[i])
    dirty[i] = False

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
                if si < TOTAL_STRIPS and pi < strip_lengths[si]:
                    off = pi * 4
                    # for GRBW strips, data must be in G,R,B,W order:
                    buffers[si][off : off + 4] = bytes((g, r, b_, w))
                    dirty[si] = True

        # drop processed bytes
        buf[:] = buf[total:]

    # push only changed strips
    for i, dio in enumerate(ios):
        if dirty[i]:
            neopixel_write(dio, buffers[i])
            dirty[i] = False
