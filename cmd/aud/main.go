package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/siggy/bbox/bbox"
)

func main() {
	files, _ := ioutil.ReadDir(bbox.WAVS)
	if len(files) != bbox.BEATS {
		panic(0)
	}

	wavs := [bbox.BEATS]*bbox.Wav{}

	for i, f := range files {
		fmt.Printf("InitWav()\n")
		wavs[i] = bbox.InitWav(f)
		// break
	}

	for _, w := range wavs {
		fmt.Printf("Play()\n")
		w.Play()
		time.Sleep(2 * time.Second)
		// break
	}

	for _, w := range wavs {
		fmt.Printf("Close()\n")
		w.Close()
		// break
	}
}

// TODO: sort this out:

// ALSA lib pulse.c:243:(pulse_connect) PulseAudio: Unable to connect: Connection refused

// Cannot connect to server socket err = No such file or directory
// Cannot connect to server request channel
// jack server is not running or cannot be started
// hihat-808.wav: 6592 samples
// OpenStream()::pStreamParameters: {Input:{Device:<nil> Channels:0 Latency:0s} Output:{Device:0x56af8900 Channels:1 Latency:34.829931ms} SampleRate:44100 FramesPerBuffer:16 Flags:0}
// InitWav()
// kick-classic.wav: 17664 samples
// OpenStream()::pStreamParameters: {Input:{Device:<nil> Channels:0 Latency:0s} Output:{Device:0x56af8900 Channels:1 Latency:34.829931ms} SampleRate:44100 FramesPerBuffer:16 Flags:0}
// Expression 'ret' failed in 'src/hostapi/alsa/pa_linux_alsa.c', line: 1736
// Expression 'AlsaOpen( &alsaApi->baseHostApiRep, params, streamDir, &self->pcm )' failed in 'src/hostapi/alsa/pa_linux_alsa.c', line: 1904
// Expression 'PaAlsaStreamComponent_Initialize( &self->playback, alsaApi, outParams, StreamDirection_Out, NULL != callback )' failed in 'src/hostapi/alsa/pa_linux_alsa.c', line: 2175
// Expression 'PaAlsaStream_Initialize( stream, alsaHostApi, inputParameters, outputParameters, sampleRate, framesPerBuffer, callback, streamFlags, userData )' failed in 'src/hostapi/alsa/pa_linux_alsa.c', line: 2840
// panic: Device unavailable

// goroutine 1 [running]:
// github.com/siggy/bbox/bbox.InitWav(0x19a058, 0x56b20240, 0x0)
// 	/home/pi/code/go/src/github.com/siggy/bbox/bbox/wav.go:66 +0x5f8
// main.main()
// 	/home/pi/code/go/src/github.com/siggy/bbox/cmd/aud/main.go:21 +0xac
// exit status 2
