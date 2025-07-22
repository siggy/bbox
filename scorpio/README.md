
## Scorpio setup

### Install CircuitPython

Based on:
https://learn.adafruit.com/introducing-feather-rp2040-scorpio/install-circuitpython

### VSCode bindings

```bash
cd ~/code/bbox/scorpio
python3 -m venv .venv
source .venv/bin/activate
python3 -m pip install --upgrade pip
pip install circuitpython-stubs
```

### Circup

```bash
pip install urllib3==1.26.15
pip install setuptools
pip install circup
```

### Copy files from mac to Scorpio

```bash
cp scorpio/lib/*   /Volumes/CIRCUITPY/lib/
cp scorpio/boot.py /Volumes/CIRCUITPY/
cp scorpio/code.py /Volumes/CIRCUITPY/
```


### Mount on Pi

```bash
sudo mkdir -p /mnt/circuitpy
sudo mount /dev/sda1 /mnt/circuitpy
ls /mnt/circuitpy

# update code
sudo cp ~/code/bbox/scorpio/code.py /mnt/circuitpy/code.py

# disconnect
sudo umount /mnt/circuitpy
```

### Debug from Pi

```bash
sudo fuser -v /dev/ttyACM0

sudo apt install picocom
picocom --baud 115200 /dev/ttyACM0
```

### Protocol

#### Physical & Serial Settings

- Connection: USB‑CDC data port (/dev/ttyACM1 on the Pi, enabled via usb_cdc.enable(data=True) on the Scorpio)
- Baud Rate: 115200 baud
- Data Bits: 8
- Parity: none
- Stop Bits: 1
- Flow Control: none

#### Packet Framing

┌────────┬────────────┬───────────────┬───────────┐
│ 0xAA   │ 2-byte LEN │   PAYLOAD     │ CHECKSUM  │
│ Start  │ Big‑endian │ (LEN bytes)   │ 1 byte    │
└────────┴────────────┴───────────────┴───────────┘

1. Start Marker: 0xAA (1 byte)
2. Length: two‑byte unsigned integer, high‑byte first, total number of payload bytes
3. Payload: a sequence of N × 6 bytes (see §3)
4. Checksum: single byte = XOR of all payload bytes

#### Payload Format

┌──────────┬──────────────┬───────┬───────┬───────┬───────┐
│ STRIP_ID │ PIXEL_INDEX  │  R    │  G    │  B    │  W    │
│  (0–7)   │ (0–length‑1) │(0–255)│(0–255)│(0–255)│(0–255)│
└──────────┴──────────────┴───────┴───────┴───────┴───────┘

- STRIP_ID (byte): Which NeoPixel strip (0…7)
- PIXEL_INDEX (byte): Which pixel on that strip
- R, G, B, W (bytes): The four color components

#### Checksum

```go
var checksum byte
for _, b := range payload {
    checksum ^= b
}
```

#### Sequence diagram

Pi (Go)                                     Scorpio (CircuitPython)
─────────                                   ────────────────────────────

 every 30 ms tick:
   ┌───────────────────┐
   │ build diff list   │
   │ serialize payload │
   └───┬───────────────┘
       │ write([0xAA…])               loop:
       │────────────────────────────▶ check in_waiting
       │                             read start byte
       │                             read length bytes
       │                             read payload
       │                             read checksum
       │                             verify checksum
       │                             for each 6‑byte group:
       │                             └─update strip[pixel]
       │                             show() on all strips
