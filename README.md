# BBox

Beatboxer in Go

## Dependencies

### portaudio

```bash
# OSX
brew install portaudio

# Raspbian
sudo apt-get install -y libasound-dev

wget http://portaudio.com/archives/pa_stable_v190600_20161030.tgz -O /tmp/pa_stable_v190600_20161030.tgz
cd /tmp
tar -xzf pa_stable_v190600_20161030.tgz
cd portaudio
./configure
make
sudo make install
sudo ldconfig
```

### rpi_ws281x

Beatboxer depends on a fork of (https://github.com/jgarff/rpi_ws281x). See that
repo for complete instructions.

```bash
sudo cp rpi_ws281x/rpihw.h  /usr/local/include/
sudo cp rpi_ws281x/ws2811.h /usr/local/include/
sudo cp rpi_ws281x/pwm.h    /usr/local/include/

sudo cp rpi_ws281x/libws2811.a /usr/local/lib/

# osx
export CGO_CFLAGS="$CGO_CFLAGS -I/usr/local/include"
export CGO_LDFLAGS="$CGO_LDFLAGS -L/usr/local/lib"
```

## Run

```bash
go build -o beatboxer cmd/bbox.go && sudo ./beatboxer
```

LED Testing

```bash
go build cmd/leds.go && sudo ./leds
```

Clear LEDs

```bash
go build cmd/clear.go && sudo ./clear
```

No LEDs

```bash
go run -race cmd/noleds.go
```

Audio Test

```bash
go run -race cmd/aud.go
```

Keyboard Test

```bash
go run -race cmd/keys.go
```

## Credits

- [wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)
- [rpi_ws281x](rpi_ws281x) courtesy of (https://github.com/jgarff/rpi_ws281x)
