package keys

import (
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	listening    bool
	keysmu       sync.Mutex
	oldtermState *terminal.State
	fd           = os.Stdin.Fd()
)

func ListenEvent() (<-chan Code, error) {
	keysmu.Lock()
	defer keysmu.Unlock()
	var err error
	if listening {
		err = errors.New("keys event on listening")
		return nil, err
	}
	if oldtermState, err = terminal.MakeRaw(int(fd)); err != nil {
		return nil, err
	}
	keycodeCh := make(chan Code)
	go func() {
		for {
			keycodeCh <- Code(getch())
		}
	}()
	listening = true
	return keycodeCh, nil
}

func StopEventListen() {
	keysmu.Lock()
	defer keysmu.Unlock()
	if !listening {
		return
	}
	if oldtermState != nil {
		terminal.Restore(0, oldtermState)
	}
	oldtermState = nil
	listening = false
}
