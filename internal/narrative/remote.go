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

type playbackSM struct {
	status playbackStatus
	mut    sync.Mutex
}

func (p *playbackSM) SetStatus(status playbackStatus) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.status = status
}

func (p *playbackSM) Status() playbackStatus {
	p.mut.Lock()
	defer p.mut.Unlock()
	return p.status
}
func (p *playbackSM) AllowsBuffering() bool {
	return p.status != playbackBuffering && p.status != playbackIdle
}

type Remote struct {
	player *player.Player
	source *Source
}

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

func (r *Remote) StartPlayback() {
	r.Play(func() {
		PlaybackCh <- 1
	})
}

func (r *Remote) Pause() {
	PSM.SetStatus(playbackPaused)
	r.player.Pause()
}

func (r *Remote) Resume() {
	PSM.SetStatus(playbackPlaying)
	r.player.Resume()
}

func (r *Remote) Stop() {
	PSM.SetStatus(playbackIdle)
	r.player.Stop()
}

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

// Initialize new remote
func NewRemote(source *Source, player *player.Player) *Remote {
	return &Remote{
		source: source,
		player: player,
	}
}
