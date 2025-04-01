package audio

import (
	"math"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/generators"
	"github.com/gopxl/beep/v2/speaker"
)

type Audio struct {
	streamer  *effects.Volume
	frequency float64
	pattern   [16]uint8
}

func New(frequency uint8) *Audio {
	a := &Audio{
		pattern:   [16]uint8{},
		frequency: 440,
	}

	sr := beep.SampleRate(48000)
	speaker.Init(sr, 4800)

	square, err := generators.SquareTone(sr, a.frequency)
	if err != nil {
		panic(err)
	}
	a.streamer = &effects.Volume{
		Streamer: square,
		Base:     2,
		Volume:   -10,
		Silent:   true,
	}

	speaker.Play(a.streamer)

	return a
}

func (a *Audio) Play() {
	if !a.streamer.Silent {
		return
	}
	a.streamer.Silent = false
	speaker.Play(a.streamer)
}

func (a *Audio) Stop() {
	if a.streamer.Silent {
		return
	}
	a.streamer.Silent = true
	speaker.Play(a.streamer)
}

func (a *Audio) LoadPattern(p [16]uint8) {
	copy(a.pattern[:], p[:])
}

func (a *Audio) SetPitch(pitch uint8) {
	a.frequency = float64(4000 * math.Exp2(float64((pitch-64)/48)))
}

func (a *Audio) Close() {
	speaker.Close()
}
