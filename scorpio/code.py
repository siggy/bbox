import board, math, time, usb_cdc, neopixel, digitalio
from neopixel import neopixel_write

ser = usb_cdc.data

# onboard heartbeat LED
hb = neopixel.NeoPixel(board.NEOPIXEL, 1, brightness=0.5, auto_write=True)

# LED strips setup
lengths = [144] * 8
pins    = [board.NEOPIXEL0, board.NEOPIXEL1, board.NEOPIXEL2, board.NEOPIXEL3,
           board.NEOPIXEL4, board.NEOPIXEL5, board.NEOPIXEL6, board.NEOPIXEL7]
ios     = []
for p in pins:
    d = digitalio.DigitalInOut(p)
    d.direction = digitalio.Direction.OUTPUT
    ios.append(d)
buffers = [bytearray(4 * L) for L in lengths]

buf = bytearray()

while True:
    # heartbeat (sine‐fade green/blue)
    t = time.monotonic()
    p = (math.sin(t * 2 * math.pi) + 1) / 2
    hb[0] = (0, int(p * 64), int((1 - p) * 64))

    # read all available bytes
    n = ser.in_waiting
    if n:
        buf.extend(ser.read(n))

    # parse 0xAA LENH LENL PAYLOAD… CHK
    while True:
        if len(buf) < 4:
            break
        if buf[0] != 0xAA:
            buf.pop(0)
            continue
        length = (buf[1] << 8) | buf[2]
        total  = 3 + length + 1
        if len(buf) < total:
            break
        payload = buf[3 : 3 + length]
        chk = buf[3 + length]
        # verify checksum
        x = 0
        for b in payload:
            x ^= b
        if x == chk and length % 6 == 0:
            # apply updates
            for i in range(0, length, 6):
                si, pi, r, g, b_, w = payload[i : i + 6]
                if si < len(buffers) and pi < lengths[si]:
                    off = pi * 4
                    buffers[si][off : off + 4] = bytes((r, g, b_, w))
        # drop this frame
        del buf[:total]

    # push to strips
    for dio, buff in zip(ios, buffers):
        neopixel_write(dio, buff)
