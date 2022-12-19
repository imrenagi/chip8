package chip8

import "C"
import (
	"math"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100

func NewAudioController() *AudioController {
	portaudio.Initialize()
	s := newStereoSine(256, 320, sampleRate)
	return &AudioController{
		stereoSine: s,
	}
}

type AudioController struct {
	stereoSine *stereoSine
	isOn       bool
}

func (a *AudioController) Destroy() {
	portaudio.Terminate()
	a.stereoSine.Close()
}

func (a *AudioController) Start() {
	if !a.isOn {
		a.isOn = true
		chk(a.stereoSine.Start())
	}
}

func (a *AudioController) Stop() {
	if a.isOn {
		a.isOn = false
		chk(a.stereoSine.Stop())
	}
}

type stereoSine struct {
	*portaudio.Stream
	stepL, phaseL float64
	stepR, phaseR float64
}

func newStereoSine(freqL, freqR, sampleRate float64) *stereoSine {
	s := &stereoSine{nil, freqL / sampleRate, 0, freqR / sampleRate, 0}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	chk(err)
	return s
}

func (g *stereoSine) processAudio(out [][]float32) {
	for i := range out[0] {
		out[0][i] = float32(math.Sin(2 * math.Pi * g.phaseL))
		_, g.phaseL = math.Modf(g.phaseL + g.stepL)
		out[1][i] = float32(math.Sin(2 * math.Pi * g.phaseR))
		_, g.phaseR = math.Modf(g.phaseR + g.stepR)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
