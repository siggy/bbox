package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/siggy/bbox/bbox2/keys"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wavs, err := wavs.Init()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer wavs.Close()

	for range 5 {
		// wavs.Play("perc-808.wav")
		// time.Sleep(250 * time.Millisecond)
		// wavs.Play("hihat-808.wav")
		// time.Sleep(250 * time.Millisecond)
		// wavs.Play("kick-classic.wav")
		// time.Sleep(250 * time.Millisecond)
		// wavs.Play("tom-808.wav")
		// time.Sleep(250 * time.Millisecond)
	}

	keys, err := keys.Init()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer keys.Close()

	go func() {
		err := keys.Run()
		if err != nil {
			log.Errorf("keys.Run failed: %v", err)
		}
	}()

	for {
		select {
		case press := <-keys.Get():
			fmt.Printf("press: %+v\n", press)
			if press.Key == keyboard.KeyCtrlC {
				log.Info("Detected Ctrl+C, exiting...")
				break
			}
			if press.Key == keyboard.KeyEsc {
				log.Info("Detected escape, exiting...")
				break
			}
		case <-ctx.Done():
			break
		}
	}
}
