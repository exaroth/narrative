package narrative

import (
	"strconv"
	"sync"

	"github.com/exaroth/narrative/pkg/player"
	"github.com/sirupsen/logrus"
)

// Playback state machine containing current status
// of playback
var PSM = playbackSM{status: playbackIdle}

// channel controlling playback.
var PlaybackCh = make(chan int)

type playbackStatus int

const (
	playbackPlaying playbackStatus = iota
	playbackPaused
	playbackBuffering
	playbackIdle
)

func (p playbackStatus) String() string {
	switch p {
	case playbackPlaying:
		return "playing"
	case playbackPaused:
		return "paused"
	case playbackBuffering:
		return "buffering"
	case playbackIdle:
		return "idle"
	}
	return "unknown"
}

// Simple state machine for managing playback
// status.
type playbackSM struct {
	// Current status
	status playbackStatus
	mut    sync.Mutex
}

// Set current playback status.
func (p *playbackSM) SetStatus(status playbackStatus) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.status = status
}

// Retrieve current playback status.
func (p *playbackSM) Status() playbackStatus {
	p.mut.Lock()
	defer p.mut.Unlock()
	return p.status
}

// Return whether we should attemptt buffering based on
// current playback status.
func (p *playbackSM) AllowsBuffering() bool {
	return p.status != playbackBuffering && p.status != playbackIdle
}

func (p *playbackSM) AllowsSourceSwitching() bool {
	return p.status != playbackBuffering
}

// Abstraction for managing playback commands for the player.
type Remote struct {
	player *player.Player
	source *Source
}

// Play current sentence, pass callback to be
// invoked when playback finishes.
func (r *Remote) Play(callback func()) {

	PSM.SetStatus(playbackPlaying)
	data, err := r.source.getCurrentSentenceWaveformData()
	if err != nil {
		// TODO
		panic(err)
	}
	r.player.AddSample(data)
	r.player.Play(callback)
}

// Initialize new playback for the source,
// It will start based on last played sentence.
func (r *Remote) StartPlayback() {
	r.Play(func() {
		PlaybackCh <- 1
	})
}

// Pause playback.
func (r *Remote) Pause() {
	PSM.SetStatus(playbackPaused)
	r.player.Pause()
}

// Resume playing current sample.
func (r *Remote) Resume() {
	PSM.SetStatus(playbackPlaying)
	r.player.Resume()
}

// Stop playback, this will remove the sample
// from, the player
func (r *Remote) Stop() {
	PSM.SetStatus(playbackIdle)
	r.player.Stop()
}

// Handle integer based playback command, and
// dispatch to the player.
func (r *Remote) HandleCommand(command int) {
	switch command {
	// continuous playback
	case 1:
		if PSM.Status() == playbackPaused {
			logrus.Info("handlePlayback: resume")
			r.Resume()
		} else {
			logrus.Info("handlePlayback: play")
			sn := r.source.incrementSentenceNum()
			// at the end of playback.
			if sn == -1 {
				// todo -> rewind
				r.Stop()
				break
			}
			r.StartPlayback()
		}
	// pause
	case 0:
		logrus.Info("handlePlayback: pause")
		r.Pause()
	// stop
	case -1:
		logrus.Info("handlePlayback: stop")
		// todo
		r.Stop()
	case 2:
		logrus.Info("handlePlayback: start playback")
		r.StartPlayback()
	default:
		panic("Invalid playback cmd passed: " + strconv.Itoa(command))
	}
}

// Initialize new remote.
func NewRemote(source *Source, player *player.Player) *Remote {
	return &Remote{
		source: source,
		player: player,
	}
}
