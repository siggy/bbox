package keys

import (
	"github.com/eiannone/keyboard"
)

type Keys struct {
	presses chan Press
}

type Press struct {
	Rune rune
	Key  keyboard.Key
}

func Init() (*Keys, error) {
	if err := keyboard.Open(); err != nil {
		return nil, err
	}

	return &Keys{presses: make(chan Press, 100)}, nil
}

func (k *Keys) Run() error {
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}
		k.presses <- Press{char, key}
	}
}

func (k *Keys) Get() <-chan Press {
	return k.presses
}

func (k *Keys) Close() error {
	return keyboard.Close()
}
