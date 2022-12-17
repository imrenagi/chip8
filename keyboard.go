package chip8

import "github.com/gdamore/tcell/v2"

type Keypad interface {
	IsPressed()
}

type Keyboard struct {
	// TODO use observer pattern
}

func (k *Keyboard) IsPressed(key uint8) bool {
	// TODO implement the keyboard
	return false
}

func (k *Keyboard) PressedEventCh() <-chan uint8 {
	ch := make(chan uint8)
	// TODO close this chan
	return ch
}

type TcellKeyboard struct {
	screen tcell.Screen
}
