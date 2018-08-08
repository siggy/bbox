package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/siggy/bbox/beatboxer/render/web"
	log "github.com/sirupsen/logrus"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	w := web.InitWeb()

	for {
		select {
		case <-sig:
			fmt.Printf("Received OS signal, exiting")
			return
		case p := <-w.Phone():
			log.Infof("Phone event: %+v", p)
		}
	}
}

// 255,0,214
// 0,255,9

// 255,255,0
// 0,0,255

// 255,0,0
// 0,255,255

// 128,0,255
// 128,255,0

// 166,0,255
// 58,255,0
