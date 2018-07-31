// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from a phone.
	phone chan []byte

	// structured inbound messages from phone
	phoneEvents chan phoneEvent

	// Outbound messages to web page
	render chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

type deviceAcceleration struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}
type deviceRotationRate struct {
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
	Gamma float64 `json:"gamma"`
}
type deviceMotion struct {
	Acceleration                 deviceAcceleration `json:"acceleration"`
	AccelerationIncludingGravity deviceAcceleration `json:"accelerationIncludingGravity"`
	Interval                     float64            `json:"interval"`
	RotationRate                 deviceRotationRate `json:"rotationRate"`
}
type deviceOrientation struct {
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
	Gamma float64 `json:"gamma"`
}
type phoneEvent struct {
	RGB string `json:"rgb"`
	r   uint64 `json:"r,omitempty"`
	g   uint64 `json:"g,omitempty"`
	b   uint64 `json:"b,omitempty"`

	Orientation deviceOrientation `json:"orientation,omitempty"`
	Motion      deviceMotion      `json:"motion,omitempty"`
}

func newHub() *Hub {
	return &Hub{
		phone:       make(chan []byte),
		render:      make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		phoneEvents: make(chan phoneEvent),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.render:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message := <-h.phone:
			p := phoneEvent{}
			err := json.Unmarshal(message, &p)
			if err != nil {
				log.Errorf("json.Unmarshal failed for %+v: %s", string(message), err)
			}
			log.Infof("PHONE EVENT MESSAGE: %+v", string(message))
			log.Infof("PHONE EVENT:         %+v", p)

			h.phoneEvents <- p
		}
	}
}
