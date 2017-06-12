# BBox

Beatboxer in Go

## Dependencies

### portaudio

```bash
brew install portaudio
```

### rpi_ws281x

Beatboxer depends on a fork of (https://github.com/jgarff/rpi_ws281x). See that
repo for complete instructions.

```bash
cp rpi_ws281x/rpihw.h  /usr/local/include/
cp rpi_ws281x/ws2811.h /usr/local/include/
cp rpi_ws281x/pwm.h    /usr/local/include/

cp rpi_ws281x/libws2811.a /usr/local/lib/
```

## Run

```bash
go build -o beatboxer cmd/bbox.go && sudo ./beatboxer
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

[wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)
[rpi_ws281x](rpi_ws281x) courtesy of (https://github.com/jgarff/rpi_ws281x)
