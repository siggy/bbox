// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"

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
	RGB              string `json:"rgb"`
	R                uint32 `json:"r,omitempty"`
	G                uint32 `json:"g,omitempty"`
	B                uint32 `json:"b,omitempty"`
	NormalizedMotion float64

	Orientation deviceOrientation `json:"orientation,omitempty"`
	Motion      deviceMotion      `json:"motion,omitempty"`
}

const (
	MIN_MAX_MOTION = 0.1

	SMOOTHING_FAST = 0.9
	SMOOTHING_SLOW = 0.99
)

var (
	motion        = float64(0)
	motionMax     = MIN_MAX_MOTION
	MAX_SMOOTHING = math.Pow(0.999, 1.0/100)

	firstRun = true

	rgbRE = regexp.MustCompile(`rgb\(([0-9]+),([0-9]+),([0-9]+)\)`)
)

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
	// rgb(1,23,121)

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
			p, err := unMarshalPhone(message)
			if err != nil {
				log.Errorf("Failed to unmarshal phone message: %+v", message)
				continue
			}

			h.phoneEvents <- p
		}
	}
}

func unMarshalPhone(message []byte) (phoneEvent, error) {
	p := phoneEvent{}

	err := json.Unmarshal(message, &p)
	if err != nil {
		log.Errorf("json.Unmarshal failed for %+v: %s", string(message), err)
		return phoneEvent{}, err
	}
	log.Debugf("Phone event message:      %+v", string(message))
	log.Debugf("Phone event unmarshalled: %+v", p)

	l := rgbRE.FindStringSubmatch(p.RGB)
	if len(l) != 4 {
		err := fmt.Errorf("Invalid rgb string: %+v", p)
		log.Error(err)
		return phoneEvent{}, err
	}

	r, err := strconv.Atoi(l[1])
	if err != nil {
		err := fmt.Errorf("Invalid rgb int parse: %+v", p)
		log.Error(err)
		return phoneEvent{}, err
	}
	g, err := strconv.Atoi(l[2])
	if err != nil {
		err := fmt.Errorf("Invalid rgb int parse: %+v", p)
		log.Error(err)
		return phoneEvent{}, err
	}
	b, err := strconv.Atoi(l[3])
	if err != nil {
		err := fmt.Errorf("Invalid rgb int parse: %+v", p)
		log.Error(err)
		return phoneEvent{}, err
	}

	p.R = uint32(r)
	p.G = uint32(g)
	p.B = uint32(b)

	p.NormalizedMotion = normalizeMotion(
		p.Motion.Acceleration.X,
		p.Motion.Acceleration.Y,
		p.Motion.Acceleration.Z,
	)

	return p, nil
}

func normalizeMotion(x, y, z float64) float64 {
	slice := []float64{x, y, z}
	bufLength := float64(len(slice))

	sum := float64(0)
	for _, n := range slice {
		x := math.Abs(float64(n) / math.MaxInt32)
		sum += math.Pow(math.Min(float64(x)/motionMax, 1), 2)
	}
	rms := math.Sqrt(sum / bufLength)

	if firstRun && rms > 0 {
		motionMax = rms
		firstRun = false
	}

	if rms > motionMax {
		motionMax = (1-SMOOTHING_FAST)*rms + motionMax*SMOOTHING_FAST
	} else {
		motionMax = (1-SMOOTHING_SLOW)*rms + motionMax*SMOOTHING_SLOW
	}

	return motionMax
}
