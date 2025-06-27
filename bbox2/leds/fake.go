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

func (f *fake) Clear() error {
	return f.Write(all(f.stripLengths))
}

func (f *fake) Write(state State) error {
	f.log.Tracef("Write: %+v", state)
	return nil
}
