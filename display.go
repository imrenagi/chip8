package chip8

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
	"github.com/rs/zerolog/log"
)

func DefaultDisplay() *Display {
	tcellScreen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal().Err(err).Msgf("cant get new tcell screen")
	}

	d := &Display{
		H:      32,
		W:      64,
		screen: tcellScreen,
	}
	d.init()
	return d
}

type Display struct {
	displayer Displayer
	screen    tcell.Screen

	H, W uint8
	data [][]uint8
}

func (d *Display) init() {
	encoding.Register()
	if err := d.screen.Init(); err != nil {
		log.Fatal().Msgf("not able to init screen")
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorReset)
	d.screen.SetStyle(defStyle)
	d.screen.Clear()

	row := make([][]uint8, d.H)
	for r := 0; r < int(d.H); r++ {
		colums := make([]uint8, d.W)
		for c := 0; c < int(d.W); c++ {
			colums[c] = 0
		}
		row[r] = colums
	}
	d.data = row

	go func() {
		for {
			switch ev := d.screen.PollEvent().(type) {
			case *tcell.EventResize:
				d.screen.Sync()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape {
					d.screen.Fini()
					os.Exit(0)
				} else if ev.Key() == tcell.KeyRune {
					fmt.Println("pressed", ev.Rune())
				}
			}
		}
	}()
}

func (d *Display) Clear() {
	d.screen.Clear()
}

func (d *Display) SetPixel(x, y uint8, val uint8) {
	d.data[y][x] = val
}

func (d *Display) GetPixel(x, y uint8) uint8 {
	return d.data[y][x]
}

func (d *Display) Draw() {
	d.screen.Clear()
	style := tcell.StyleDefault.Background(tcell.ColorDarkRed)
	for y, r := range d.data {
		for x, c := range r {
			if c > 0 {
				emitStr(d.screen, x, y, style, " ")
			}
		}
	}
	d.screen.Show()
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

type Displayer interface {
	Clear()
	Draw()
}

type TcellDisplay struct {
}

func (t TcellDisplay) Draw() {

}

func (t TcellDisplay) Clear() {

}

func (t TcellDisplay) Show() {

}
