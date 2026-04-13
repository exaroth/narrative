package player

import (
	"github.com/gopxl/beep/v2"
)

// Return new controller wrapped sample,
// allowing for stream pausing.
func NewSampleCtrl(data []float32) *beep.Ctrl {
	return &beep.Ctrl{
		Streamer: NewSample(data),
		Paused:   true,
	}
}

// Initialize new Sample with data
func NewSample(data []float32) beep.Streamer {
	return &Sample{data, 0}
}

// Representation of sample to be streamed.
type Sample struct {
	// Raw data
	data []float32
	// Index of last processed chunk
	processed uint
}

// Implements streaming method for the sample.
func (s *Sample) Stream(samples [][2]float64) (n int, ok bool) {
	var t_sample float64

	d := s.data[s.processed:]

	for idx, sample := range d {
		if idx > len(samples)-1 {
			break
		}
		t_sample = float64(sample)
		samples[idx][1] = t_sample
		samples[idx][0] = t_sample
	}
	processed := int(s.processed) + len(samples)
	if processed > len(s.data) {
		return 0, false
	}
	s.processed = uint(processed)
	return len(samples), true
}

func (*Sample) Err() error {
	return nil
}
