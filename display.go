package chip8

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

func DefaultDisplay() Display {
	d := Display{
		H:      64,
		W:      128,
		drawer: NewSDLDisplay(),
	}
	return d
}

type Display struct {
	drawer Drawer

	H, W uint8
	data [8192]uint8
}

func (d *Display) Clear() {
	d.drawer.Clear()
}

func (d *Display) SetPixel(x, y uint8, val uint8) {
	t := uint16(y)*uint16(d.W) + uint16(x)
	d.data[t] = val
}

func (d *Display) GetPixel(x, y uint8) uint8 {
	t := uint16(y)*uint16(d.W) + uint16(x)
	return d.data[t]
}

func (d *Display) Draw() {
	d.drawer.Clear()
	for i, val := range d.data {
		y := i / int(d.W) // 4 / 3 = 1
		x := i % int(d.W) // 4 % 3 = 2
		if val > 0 {
			d.drawer.SetPixel(x, y)
		}
	}
	d.drawer.Draw()
}

func (d *Display) Stop() {
	d.drawer.Stop()
}

type Drawer interface {
	Clear()
	SetPixel(x, y int)
	Draw()
	Stop()
}

func NewSDLDisplay() *SDLDisplay {
	window, err := sdl.CreateWindow("chip-8 emulator",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		1280, 640, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	return &SDLDisplay{
		window:  window,
		surface: surface,
	}
}

type SDLDisplay struct {
	sync.Mutex
	window  *sdl.Window
	surface *sdl.Surface
}

func (s *SDLDisplay) Clear() {
	s.surface.FillRect(nil, 0)
}

func (s *SDLDisplay) SetPixel(x, y int) {
	rect := sdl.Rect{int32(x) * 10, int32(y) * 10, 10, 10}
	s.surface.FillRect(&rect, 0xffff0000)
}

func (s *SDLDisplay) Draw() {
	s.window.UpdateSurface()
}

func (s *SDLDisplay) Stop() {
	s.window.Destroy()
}
