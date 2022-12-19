package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/imrenagi/chip8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)

	go func() {
		oscall := <-ch
		log.Warn().Msgf("system call:%+v", oscall)
		cancel()
	}()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	display := chip8.DefaultDisplay()
	keyboard := chip8.NewKeyboard()
	audio := chip8.NewAudioController()

	c := chip8.NewCPU(display, keyboard, audio)
	c.LoadProgram("examples/c8games/PONG")
	go c.Start(ctx)

exit:
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				break exit
			case *sdl.KeyboardEvent:
				ke := event.(*sdl.KeyboardEvent)
				var pressed bool
				if ke.State == sdl.PRESSED {
					pressed = true
				}
				keyboard.Accept(chip8.NewKeyEvent(pressed, ke.Keysym.Scancode))
			}
		}
	}
}
