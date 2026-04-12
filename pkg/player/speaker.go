package player

import (
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

var (
	// Default samplerate for KittenTTS
	// models - should not be changed
	SampleRate = 24000
	// Default size of the buffer to be streamed.
	BufferSize = 2400
)

// Thin wrapper over beep Speaker module
// for handling sample playback.
type Player struct {
	ctrl *beep.Ctrl
}

// Initialize new Player
func InitPlayer() *Player {
	sr := beep.SampleRate(SampleRate)
	speaker.Init(sr, BufferSize)
	return &Player{
		ctrl: &beep.Ctrl{
			Streamer: nil,
			Paused:   true,
		},
	}
}

// Add new sample to player's controller.
// Assumes waveform is in native []float32
// format.
func (p *Player) AddSample(data []float32) {
	p.ctrl.Streamer = NewSample(data)
}

// Play currently loaded sample (if present).
func (p *Player) Play() {
	if p.ctrl.Streamer == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = false
	speaker.Unlock()
	speaker.Play(p.ctrl)
}

// Stop playing current sample.
func (p *Player) Pause() {
	if p.ctrl.Streamer == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = true
	speaker.Unlock()
}

// Toggle playback of current sample.
func (p *Player) Toggle() {
	if p.ctrl.Streamer == nil {
		return
	}
	speaker.Lock()
	if p.ctrl.Paused == true {
		p.Play()
	} else {
		p.Pause()
	}
}
