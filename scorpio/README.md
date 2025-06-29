
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
