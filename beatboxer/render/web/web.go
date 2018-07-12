package web

import (
	"fmt"
	"log"
	"net/http"
)

type Web struct {
	hub *Hub
}

func InitWeb() *Web {
	fmt.Printf("InitWeb\n")

	hub := newHub()
	go hub.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "beatboxer/render/web/index.html")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	return &Web{
		hub: hub,
	}
}

func (w *Web) Init(
	freq int,
	gpioPin1 int, ledCount1 int, brightness1 int,
	gpioPin2 int, ledCount2 int, brightness2 int,
) error {
	return nil
}

func (w *Web) Fini() {

}

func (w *Web) Render() error {
	return nil
}

func (w *Web) Wait() error {
	return nil
}

func (w *Web) SetLed(channel int, index int, value uint32) {
	w.hub.send(fmt.Sprintf("%d: %3d %d", channel, index, value))
}

func (w *Web) Clear() {

}

func (w *Web) SetBitmap(channel int, a []uint32) {

}
