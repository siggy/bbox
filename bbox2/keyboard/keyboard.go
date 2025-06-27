package keyboard

import (
	"github.com/eiannone/keyboard"
	"github.com/siggy/bbox/bbox2/program"
	log "github.com/sirupsen/logrus"
)

type Keyboard struct {
	keymaps   map[rune]program.Coord
	keyEvents <-chan keyboard.KeyEvent
	presses   chan program.Coord

	log *log.Entry
}

const keyBuffer = 100

// runeMap overrides the two special keys from MacroPaw Beatboxer
var runeMap = map[keyboard.Key]rune{
	keyboard.KeySpace: '8',
	keyboard.KeyEnter: '9',
}

func New(keymaps map[rune]program.Coord) (*Keyboard, error) {
	keyEvents, err := keyboard.GetKeys(keyBuffer)
	if err != nil {
		return nil, err
	}

	return &Keyboard{
		keymaps:   keymaps,
		keyEvents: keyEvents,
		presses:   make(chan program.Coord, keyBuffer),

		log: log.WithField("bbox2", "keyboard"),
	}, nil
}

func (k *Keyboard) Presses() <-chan program.Coord {
	return k.presses
}

func (k *Keyboard) Run() {
	defer func() {
		keyboard.Close()
		close(k.presses)
	}()

	for {
		event := <-k.keyEvents
		if event.Err != nil {
			return
		}
		k.log.Debugf("You pressed: rune %q, key %X", event.Rune, event.Key)

		switch event.Key {
		case keyboard.KeyEsc:
			k.log.Info("Detected Escape, exiting...")
			return
		case keyboard.KeyCtrlC:
			k.log.Info("Detected Ctrl+C, exiting...")
			return
		}

		r := event.Rune
		if r == 0 {
			ok := false
			r, ok = runeMap[event.Key]
			if !ok {
				k.log.Warnf("Key %X has no mapped rune, skipping", event.Key)
				continue
			}
		}

		coord, ok := k.keymaps[r]
		if !ok {
			k.log.Warnf("No coordinates for key %q", r)
			continue
		}

		k.presses <- coord
	}
}
