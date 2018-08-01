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
			log.Debugf("Phone event: %+v", p)
		}
	}
}
