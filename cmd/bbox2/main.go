package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siggy/bbox/bbox2/keys"
	"github.com/siggy/bbox/bbox2/wavs"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wavs, err := wavs.Init()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer wavs.Close()

	for range 2 {
		wavs.Play("perc-808.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("hihat-808.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("kick-classic.wav")
		time.Sleep(250 * time.Millisecond)
		wavs.Play("tom-808.wav")
		time.Sleep(250 * time.Millisecond)
	}

	keys, err := keys.Init()
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}

	presses := keys.Run()

	for {
		select {
		case press, more := <-presses:
			if !more {
				log.Info("key channel closed, exiting...")
				return
			}
			log.Debugf("press: %q", press)
		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}
