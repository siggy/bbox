package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func main() {
	// Read the mp3 file into memory
	fileBytes, err := os.ReadFile("./wavs/ceottk001_human.wav")
	if err != nil {
		panic("reading my-file.mp3 failed: " + err.Error())
	}

	// // Convert the pure bytes into a reader object that can be used with the mp3 decoder
	// fileBytesReader := bytes.NewReader(fileBytes)

	// Decode file
	reader := wav.NewDecoder(bytes.NewReader(fileBytes))
	if !reader.IsValidFile() {
		log.Fatal("Invalid WAV file")
	}

	// reader.
	buf, err := reader.FullPCMBuffer()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Audio format code: %d\n", reader.WavAudioFormat)
	fmt.Printf("WAV file format: %v\n", buf.Format)

	if buf.Format.SampleRate == 0 || buf.Format.NumChannels == 0 {
		log.Fatal("Missing WAV format information")
	}

	pcm := intBufferTo16LEBytes(buf)

	// Prepare an Oto context (this will use your default audio device) that will
	// play all our sounds. Its configuration can't be changed later.

	// Remember that you should **not** create more than one context
	// Setup Oto audio context
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   buf.Format.SampleRate,
		ChannelCount: buf.Format.NumChannels,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		log.Fatalf("Failed to create Oto context: %v", err)
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-ready

	// Create and play audio player
	player := ctx.NewPlayer(bytes.NewReader(pcm))
	player.Play()

	for player.IsPlaying() {
		// fmt.Printf("Player is playing: %v\n", player.IsPlaying())
		time.Sleep(time.Millisecond)
	}

	// intBuf := buf.AsIntBuffer()
	// intBuf.SourceBitDepth = 32

	// // Convert 32-bit PCM to 16-bit PCM by shifting
	// for i, sample := range buf.Data {
	// 	// Shift from 32-bit to 16-bit signed (simple scaling)
	// 	buf.Data[i] = sample >> 16
	// }

	// // Now it's safe to encode as 16-bit PCM bytes
	// raw := intBufferToFloat32LEBytes(buf)
	// intBuffer := buf.AsIntBuffer()
	// transforms.PCMScaleF32()

	// // Create a player from a bytes.Reader
	// player := otoCtx.NewPlayer(bytes.NewReader(nil))
	// // player.Play()

	// // We can wait for the sound to finish playing using something like this
	// for player.IsPlaying() {
	// 	// fmt.Printf("Player is playing: %v\n", player.IsPlaying())
	// 	time.Sleep(time.Millisecond)
	// }
	// If you don't want the player/sound anymore simply close
	err = player.Close()
	if err != nil {
		panic("player.Close failed: " + err.Error())
	}
}

// Converts 32-bit integer PCM buffer to float32 little-endian bytes for oto.FormatFloat32LE
func intBufferToFloat32LEBytes(buf *audio.IntBuffer) []byte {
	out := make([]byte, len(buf.Data)*4) // 4 bytes per float32 sample
	for i, sample := range buf.Data {
		// Convert int32 sample (assumed full-scale) to float32 in [-1, 1]
		f32 := float32(sample) / float32(math.MaxInt32)
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(f32))
	}
	return out
}

// Converts a 16-bit IntBuffer to little-endian PCM []byte
func intBufferTo16LEBytes(buf *audio.IntBuffer) []byte {
	out := make([]byte, len(buf.Data)*2)
	for i, sample := range buf.Data {
		// Clamp sample to int16 range (optional safety)
		if sample > 32767 {
			sample = 32767
		} else if sample < -32768 {
			sample = -32768
		}
		binary.LittleEndian.PutUint16(out[i*2:], uint16(int16(sample)))
	}
	return out
}
