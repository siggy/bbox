package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
		case key := <-keys.Get():
			fmt.Printf("key: %q\n", key)
			// if key == keyboard.KeyEsc {
			// 	break
			// }
		case <-ctx.Done():
			return
		}
	}
}
