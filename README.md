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

## TODO

- graceful shutdown

    ```golang
    stream.Stop()
    stream.Close()
    ````
