package narrative

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/player"
	"github.com/orcaman/writerseeker"
	"github.com/viert/go-lame"
)

const ENCODER_SENTENCES_PER_FILE = 20
const ENCODER_TEMP_DIR = "/tmp/narrative-encoder"

// Convert input source to mp3.
func (c *NarrativeCtrl) Convert() (msg string, err error) {
	if len(c.args.Source) == 0 {
		err = fmt.Errorf("Must pass path to text source in argument")
		return
	}

	lib_path, err := c.dataCfg.GetLibPath(c.paths)
	if err != nil {
		return
	}

	c.ttsClient = kitten.InitKittenWithParams(lib_path,
		c.dataCfg.GetModelPath(c.paths),
		c.dataCfg.GetVoicesPath(c.paths),
		c.cfg.Voice,
		c.cfg.Speed,
	)

	if err = c.addNewSource(); err != nil {
		return
	}

	if err = c.LoadSource(c.autoplaySource); err != nil {
		return
	}

	speaker := player.InitPlayer()

	if err = os.MkdirAll(ENCODER_TEMP_DIR, os.ModePerm); err != nil {
		err = fmt.Errorf("Error creating temp dir for encoder %w", err)
		return
	}
	// defer os.RemoveAll(ENCODER_TEMP_DIR)

	var (
		eos              bool
		chunk_i          int
		chunk_f, chunk_p string
		chunk_fa         []string
	)

LOOP:
	for {
		var wd []float32
		ws := &writerseeker.WriterSeeker{}
		for _ = range ENCODER_SENTENCES_PER_FILE {
			this_wd, e := c.currentSource.getCurrentSentenceWaveformData()
			if e != nil {
				err = e
				return
			}
			wd = append(wd, this_wd...)
			if sn := c.currentSource.IncrementSentenceNum(); sn == -1 {
				eos = true
				break
			}
		}
		SpinnerMessageCh <- fmt.Sprintf(
			"Converting text source %3.0f%%",
			float64(c.currentSource.SNum())/float64(c.currentSource.Length())*100,
		)

		if len(wd) == 0 {
			break LOOP
		}

		speaker.AddSample(wd)
		if err = speaker.EncodeWav(ws); err != nil {
			return
		}

		chunk_f = fmt.Sprintf("chunk-%d.mp3", chunk_i)
		chunk_p = filepath.Join(ENCODER_TEMP_DIR, chunk_f)
		if err = processMp3FileChunk(chunk_p, ws.BytesReader()); err != nil {
			return
		}
		chunk_fa = append(chunk_fa, chunk_f)

		if eos {
			break LOOP
		}

		chunk_i += 1
	}

	return "", nil
}

// Create single mp3 chunk from wav data at path specified.
func processMp3FileChunk(chunk_path string, data *bytes.Reader) error {
	var err error
	of, err := os.Create(chunk_path)
	if err != nil {
		return err
	}
	defer of.Close()
	enc := getLameEncoder(of)
	defer enc.Close()
	_, err = data.WriteTo(enc)
	return err
}

// Retrieve lame encoder set up with mp3 options.
func getLameEncoder(f *os.File) *lame.Encoder {
	enc := lame.NewEncoder(f)
	enc.SetInSamplerate(24000)
	enc.SetNumChannels(2)
	enc.SetLowPassFrequency(-1)
	enc.SetQuality(5)
	// mono
	enc.SetMode(3)
	return enc
}
