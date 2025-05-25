package keys

import (
	"github.com/eiannone/keyboard"
	log "github.com/sirupsen/logrus"
)

const keyBuffer = 100

type Keys struct {
	keyEvents <-chan keyboard.KeyEvent
}

func Init() (*Keys, error) {
	keyEvents, err := keyboard.GetKeys(keyBuffer)
	if err != nil {
		return nil, err
	}

	return &Keys{keyEvents: keyEvents}, nil
}

func (k *Keys) Run() <-chan rune {
	ch := make(chan rune, keyBuffer)

	go func() {
		defer func() {
			keyboard.Close()
			close(ch)
		}()

		for {
			event := <-k.keyEvents
			if event.Err != nil {
				return
			}
			log.Debugf("You pressed: rune %q, key %X", event.Rune, event.Key)

			switch event.Key {
			case keyboard.KeyEsc:
				log.Debug("Detected Escape, exiting...")
				return
			case keyboard.KeyCtrlC:
				log.Debug("Detected Ctrl+C, exiting...")
				return
			}

			ch <- event.Rune
		}
	}()

	return ch
}
