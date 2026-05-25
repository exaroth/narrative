package player

import (
	"fmt"
	"io"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
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
	ctrl   *beep.Ctrl
	format *beep.Format
}

// Initialize new Player
func InitPlayer() *Player {
	sr := beep.SampleRate(SampleRate)
	speaker.Init(sr, BufferSize)
	return &Player{
		format: &beep.Format{
			SampleRate:  sr,
			NumChannels: 2,
			Precision:   2,
		},
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
func (p *Player) Play(callback func()) {
	if p.ctrl.Streamer == nil {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = false

	sounds := []beep.Streamer{
		p.ctrl,
	}
	if callback != nil {
		sounds = append(sounds, beep.Callback(callback))
	}
	speaker.Unlock()

	speaker.Play(beep.Seq(sounds...))
}

func (p *Player) EncodeWav(w io.WriteSeeker) error {
	if p.ctrl.Streamer == nil {
		return fmt.Errorf("Attempting to encode nil streamer")
	}
	return wav.Encode(w, p.ctrl.Streamer, *p.format)
}

// Pause playback of current sample.
func (p *Player) Pause() {
	if p.ctrl.Streamer == nil {
		return
	}

	speaker.Lock()
	defer speaker.Unlock()
	p.ctrl.Paused = true
}

func (p *Player) Resume() {
	if p.ctrl.Streamer == nil {
		return
	}

	speaker.Lock()
	defer speaker.Unlock()
	p.ctrl.Paused = false

}

// Stop playing current sample, this also
// removes it from the ctrl.
func (p *Player) Stop() {
	speaker.Lock()
	p.ctrl.Streamer = nil
	speaker.Unlock()
	speaker.Clear()
}

// Toggle playback of current sample.
func (p *Player) Toggle() {
	if p.ctrl.Streamer == nil {
		return
	}
	speaker.Lock()
	defer speaker.Unlock()
	p.ctrl.Paused = !p.ctrl.Paused
}
