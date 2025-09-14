
# SCORPIO setup

## Configure board for Arduino

1. Arduino > Preferences
2. Add https://github.com/earlephilhower/arduino-pico/releases/download/global/package_rp2040_index.json
3. Boards Manager -> Raspberry Pi Pico/RP2040
4. Boards -> Adafruit Feather RP2040 SCORPIO
5. Sketch -> Include Library -> Manage Libraries ->
  - Adafruit_NeoPXL8
  - Adafruit_NeoPixel
  - Adafruit_ZeroDMA

## Arduino programs for SCORPIO

- [bbox.ino](bbox.ino): Main Beatboxer program.
- [baux.ino](baux.ino): Auxilliary Beatboxer program for driving the base and globe.
- [strand.ino](strand.ino): Standalone NeoPixel strand experiment.

## Debug from Pi

```bash
sudo fuser -v /dev/ttyACM0

sudo apt install picocom
picocom --baud 115200 /dev/ttyACM0
```

## Protocol

### Physical & Serial Settings

- Connection: USB‑CDC data port (/dev/ttyACM1 on the Pi, enabled via usb_cdc.enable(data=True) on the SCORPIO)
- Baud Rate: 115200 baud
- Data Bits: 8
- Parity: none
- Stop Bits: 1
- Flow Control: none

### Packet Framing

```
┌────────┬────────────┬───────────────┬───────────┐
│ 0xAA   │ 2-byte LEN │   PAYLOAD     │ CHECKSUM  │
│ Start  │ Big‑endian │ (LEN bytes)   │ 1 byte    │
└────────┴────────────┴───────────────┴───────────┘
```

1. Start Marker: 0xAA (1 byte)
2. Length: two‑byte unsigned integer, high‑byte first, total number of payload bytes
3. Payload: a sequence of N × 6 bytes (see §3)
4. Checksum: single byte = XOR of all payload bytes

### Payload Format

```
┌──────────┬──────────────┬───────┬───────┬───────┬───────┐
│ STRIP_ID │ PIXEL_INDEX  │  R    │  G    │  B    │  W    │
│  (0–7)   │ (0–length‑1) │(0–255)│(0–255)│(0–255)│(0–255)│
└──────────┴──────────────┴───────┴───────┴───────┴───────┘
```

- STRIP_ID (byte): Which NeoPixel strip (0…7)
- PIXEL_INDEX (byte): Which pixel on that strip
- R, G, B, W (bytes): The four color components

### Checksum

```go
var checksum byte
for _, b := range payload {
    checksum ^= b
}
```

### Sequence diagram

```
Pi (Go)                               SCORPIO
───────                               ───────

 every 30 ms tick:
   ┌───────────────────┐
   │ build diff list   │
   │ serialize payload │
   └───┬───────────────┘
       │ write([0xAA…])               loop:
       │────────────────────────────▶ check in_waiting
       │                              read start byte
       │                              read length bytes
       │                              read payload
       │                              read checksum
       │                              verify checksum
       │                              for each 6‑byte group:
       │                              └─update strip[pixel]
       │                              show() on all strips
```
