package leds

import (
	log "github.com/sirupsen/logrus"
)

type fake struct {
	stripLengths []int
	log          *log.Entry
}

func NewFake(stripLengths []int) (LEDs, error) {
	log := log.WithField("leds", "fake")

	log.Infof("NewFake: %+v", stripLengths)

	return &fake{
		stripLengths: stripLengths,
		log:          log,
	}, nil
}

func (f *fake) Close() error {
	return nil
}

func (f *fake) Clear() {
	f.Set(all(f.stripLengths))
}

func (f *fake) Set(state State) {
	f.log.Tracef("Set: %+v", state)
}
