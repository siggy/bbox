# Developing Beatboxer Programs

This document describes development of "programs" for the Beatboxer.

## Quick Start

From the root of this repo:

```bash
go run cmd/beatboxer_noleds.go
```

Optional web renderer:

```bash
open http://localhost:8080
```

The Beatboxer's 16x4 grid of buttons is emulated via your keyboard. The first
8x4 keys are bound to:
- `1` .. `8`
- `q` .. `i`
- `a` .. `k`
- `z` .. `,`

The shift key modifies the above to act on the second set of 8x4 keys.

Note all indexes in code are 0-based. Button 0x0 is `1`, 15x3 is `shift` + `,`.

Pressing button 15x1 (`shift` + `i`) 5 times will switch programs.

## Adding a Program

1. Add a new file under [`programs`](programs/)
1. Implement the `Program` interface, as defined in [`program.go`](program.go).
    This interface supports the following:
    - Input
        - `New`: initialize state, kicks off long-running processes, returns new handle
        - `Amplitude`: receive amplitude, [0..1]
        - `Keyboard`: receive button presses
        - `Close`: stops all processes, frees resources
    - Output
        - `Play`: play audio files
        - `Render`: render LEDs
        - `Yield`: yield back to the harness

1. Add your new module to [`../cmd/beatboxer_noleds.go`](../cmd/beatboxer_noleds.go).

## TODO

- update LED rendering code for harness environment
- consider buffered channels, drop when channels fills
