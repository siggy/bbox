package keys

import (
	"github.com/eiannone/keyboard"
)

type Keys struct {
	presses chan rune
}

func Init() (*Keys, error) {
	if err := keyboard.Open(); err != nil {
		return nil, err
	}

	return &Keys{presses: make(chan rune, 100)}, nil
}

func (k *Keys) Run() error {
	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			return err
		}
		k.presses <- char
	}
}

func (k *Keys) Get() <-chan rune {
	return k.presses
}

func (k *Keys) Close() error {
	return keyboard.Close()
}
