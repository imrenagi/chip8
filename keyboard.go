package chip8

import (
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Keypad interface {
	IsPressed()
}

var (
	keyMap = map[rune]uint8{
		'x': 0x0, // 0
		'1': 0x1, // 1
		'2': 0x2, // 2
		'3': 0x3, // 3
		'q': 0x4, // 4
		'w': 0x5, // 5
		'e': 0x6, // 6
		'a': 0x7, // 7
		's': 0x8, // 8
		'd': 0x9, // 9
		'z': 0xA, // A
		'c': 0xB, // B
		'4': 0xC, // C
		'r': 0xD, // D
		'f': 0xE, // E
		'v': 0xF, // F
	}
)

func newKey(r rune) *key {
	k := &key{
		r:       r,
		eventCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
	}

	go k.observe()
	return k
}

type key struct {
	r            rune
	beingPressed bool
	eventCh      chan struct{}
	stopCh       chan struct{}
}

func (k *key) observe() {
	for {
		select {
		case <-k.eventCh:
			k.beingPressed = true
		case <-time.After(200 * time.Millisecond):
			k.beingPressed = false
		case <-k.stopCh:
			return
		}
	}
}

func (k *key) stop() {
	close(k.stopCh)
}

func NewKeyboard(screen tcell.Screen) *Keyboard {
	k := &Keyboard{
		screen:  screen,
		quitCh:  make(chan struct{}),
		eventCh: make(chan uint8),
	}

	for s, v := range keyMap {
		k.keys[v] = newKey(s)
	}

	eventCh := make(chan tcell.Event)
	go screen.ChannelEvents(eventCh, k.quitCh)

	go func() {
		for {
			select {
			case event := <-eventCh:
				switch ev := event.(type) {
				case *tcell.EventResize:
					screen.Sync()
				case *tcell.EventKey:
					if ev.Key() == tcell.KeyEscape {
						screen.Fini()
						os.Exit(0)
					} else if ev.Key() == tcell.KeyRune {
						k.keys[keyMap[ev.Rune()]].eventCh <- struct{}{}
						select {
						case k.eventCh <- keyMap[ev.Rune()]:
						default:
						}
					}
				}
			case <-k.quitCh:
				return
			}
		}
	}()

	return k
}

type Keyboard struct {
	sync.Mutex

	screen tcell.Screen

	keys [16]*key

	quitCh  chan struct{}
	eventCh chan uint8
}

func (k *Keyboard) IsBeingPressed(key uint8) bool {
	k.Lock()
	defer k.Unlock()
	return k.keys[key].beingPressed
}

func (k *Keyboard) PressedEventCh() <-chan uint8 {
	return k.eventCh
}
