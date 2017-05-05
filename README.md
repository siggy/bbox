# BBox

Beatboxer in Go

## Dependencies

```bash
brew install portaudio
```

## Run

```bash
go run -race cmd/bbox/main.go
```

## Credits

[wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)

## TODO

```
WARNING: DATA RACE
Read at 0x0000042795e0 by goroutine 11:
  github.com/siggy/bbox/vendor/github.com/nsf/termbox-go.Flush()
      /Users/sig/code/go/src/github.com/siggy/bbox/vendor/github.com/nsf/termbox-go/api.go:196 +0x6ac
  github.com/siggy/bbox/bbox.(*Render).Run()
      /Users/sig/code/go/src/github.com/siggy/bbox/bbox/render.go:50 +0x19d

Previous write at 0x0000042795e0 by goroutine 12:
  github.com/siggy/bbox/vendor/github.com/nsf/termbox-go.Close()
      /Users/sig/code/go/src/github.com/siggy/bbox/vendor/github.com/nsf/termbox-go/api.go:143 +0x533
  github.com/siggy/bbox/bbox.(*Keyboard).Emitter()
      /Users/sig/code/go/src/github.com/siggy/bbox/bbox/keyboard.go:173 +0x21f

Goroutine 11 (running) created at:
  main.main()
      /Users/sig/code/go/src/github.com/siggy/bbox/cmd/bbox/main.go:26 +0x289

Goroutine 12 (finished) created at:
  github.com/siggy/bbox/bbox.(*Keyboard).Run()
      /Users/sig/code/go/src/github.com/siggy/bbox/bbox/keyboard.go:112 +0x119
```
