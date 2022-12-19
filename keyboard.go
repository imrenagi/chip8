package chip8

import (
	"fmt"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	keyMap = map[sdl.Scancode]uint8{
		sdl.SCANCODE_X: 0x0, // 0
		sdl.SCANCODE_1: 0x1, // 1
		sdl.SCANCODE_2: 0x2, // 2
		sdl.SCANCODE_3: 0x3, // 3
		sdl.SCANCODE_Q: 0x4, // 4
		sdl.SCANCODE_W: 0x5, // 5
		sdl.SCANCODE_E: 0x6, // 6
		sdl.SCANCODE_A: 0x7, // 7
		sdl.SCANCODE_S: 0x8, // 8
		sdl.SCANCODE_D: 0x9, // 9
		sdl.SCANCODE_Z: 0xA, // A
		sdl.SCANCODE_C: 0xB, // B
		sdl.SCANCODE_4: 0xC, // C
		sdl.SCANCODE_R: 0xD, // D
		sdl.SCANCODE_F: 0xE, // E
		sdl.SCANCODE_V: 0xF, // F
	}
)

func NewKeyboard() *Keyboard {
	k := &Keyboard{
		quitCh:  make(chan struct{}),
		eventCh: make(chan KeyEvent, 10),
		testCh:  make(chan uint8),
	}
	go k.Observe()
	return k
}

type Keyboard struct {
	sync.Mutex
	
	keyState [16]bool

	quitCh  chan struct{}
	eventCh chan KeyEvent
	testCh  chan uint8
}

func (k *Keyboard) Observe() {
	for {
		select {
		case ev := <-k.eventCh:
			k.process(ev)
		case <-k.quitCh:
			return
		}
	}
}

type KeyEvent struct {
	Pressed  bool
	ScanCode sdl.Scancode
}

func (k *Keyboard) process(ev KeyEvent) {
	idx, ok := keyMap[ev.ScanCode]
	if !ok {
		return
	}
	fmt.Println(ev.Pressed, ev.ScanCode, idx)
	if ev.Pressed {
		k.keyState[idx] = true
		select {
		case k.testCh <- idx:
		default:
		}
	} else {
		k.keyState[idx] = false
	}
}

func (k *Keyboard) Accept(ev KeyEvent) {
	k.eventCh <- ev
}

func (k *Keyboard) IsBeingPressed(key uint8) bool {
	k.Lock()
	defer k.Unlock()
	return k.keyState[key]
}

func (k *Keyboard) PressedEventCh() <-chan uint8 {
	return k.testCh
}
