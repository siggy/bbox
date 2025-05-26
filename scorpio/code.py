# SPDX-FileCopyrightText: 2021 Kattni Rembor for Adafruit Industries
#
# SPDX-License-Identifier: MIT

"""
Blink example for boards with ONLY a NeoPixel LED (e.g. without a built-in red LED).
Includes QT Py and various Trinkeys.

Requires two libraries from the Adafruit CircuitPython Library Bundle.
Download the bundle from circuitpython.org/libraries and copy the
following files to your CIRCUITPY/lib folder:
* neopixel.mpy
* adafruit_pixelbuf.mpy

Once the libraries are copied, save this file as code.py to your CIRCUITPY
drive to run it.
"""
import time
import board
import neopixel

pixels = neopixel.NeoPixel(
    board.NEOPIXEL0,
    30,
    pixel_order=neopixel.GRBW,
    bpp=4,
    auto_write=False,
    brightness=1.0
)

while True:
    for i in range(len(pixels)):
        pixels.fill((0, 0, 0, 0))           # turn all off
        pixels[i] = (10, 0, 0, 0)           # red
        pixels.show()
        time.sleep(0.05)

    for i in range(len(pixels)):
        pixels.fill((0, 0, 0, 0))
        pixels[i] = (0, 10, 0, 0)           # green
        pixels.show()
        time.sleep(0.05)

    for i in range(len(pixels)):
        pixels.fill((0, 0, 0, 0))
        pixels[i] = (0, 0, 10, 0)           # blue
        pixels.show()
        time.sleep(0.05)

    for i in range(len(pixels)):
        pixels.fill((0, 0, 0, 0))
        pixels[i] = (0, 0, 0, 10)           # white
        pixels.show()
        time.sleep(0.05)
