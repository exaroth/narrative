package narrative

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/braheezy/shine-mp3/pkg/mp3"
	"github.com/exaroth/narrative/pkg/kitten"
	"github.com/exaroth/narrative/pkg/player"
	"github.com/go-audio/wav"
	"github.com/hyacinthus/mp3join"
	"github.com/orcaman/writerseeker"
)

const (
	ENCODER_SENTENCES_PER_FILE = 20
	ENCODER_TEMP_DIR           = "/tmp/narrative-encoder"
	ENCODER_OUTPUT_FILE        = "narrative.mp3"
)

// Convert input source to mp3.
func (c *NarrativeCtrl) Convert() (err error) {
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
			"Converting text source...%3.0f%%",
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
		if err = processMp3FileChunk(
			chunk_p,
			ws.BytesReader(),
			c.currentSource,
		); err != nil {
			return
		}
		chunk_fa = append(chunk_fa, chunk_f)

		if eos {
			break LOOP
		}

		chunk_i += 1
	}

	mp3Joiner := mp3join.New()
	// Merge mp3 chunks into final file
	for idx, f := range chunk_fa {
		SpinnerMessageCh <- fmt.Sprintf(
			"Merging files...%3.0f%%",
			float64(idx)/float64(len(chunk_fa))*100,
		)
		chunk_p = filepath.Join(ENCODER_TEMP_DIR, f)
		file, err := os.Open(chunk_p)
		if err != nil {
			return err
		}
		mp3Joiner.Append(file)
	}

	output_f, err := os.Create(ENCODER_OUTPUT_FILE)
	if err != nil {
		err = fmt.Errorf("Error creating output file: %w", err)
		return
	}
	defer output_f.Close()

	writer := bufio.NewWriter(output_f)
	reader := bufio.NewReader(mp3Joiner.Reader())
	buf := make([]byte, 10*1024)
	SpinnerMessageCh <- "Saving mp3 file..."
	for {
		v, _ := reader.Read(buf)
		writer.Write(buf)
		if v == 0 {
			break
		}
	}
	writer.Flush()
	return nil
}

// Create single mp3 chunk from wav data at path specified.
func processMp3FileChunk(chunk_path string, data *bytes.Reader, s *Source) error {
	var err error
	of, err := os.Create(chunk_path)
	if err != nil {
		return err
	}
	defer of.Close()

	wavDecoder := wav.NewDecoder(data)
	wavBuffer, err := wavDecoder.FullPCMBuffer()
	if err != nil {
		log.Fatalf("Error decoding WAV file: %v", err)
	}

	decodedData := make([]int16, len(wavBuffer.Data))
	for i, val := range wavBuffer.Data {
		decodedData[i] = int16(val)
	}

	mp3Encoder := mp3.NewEncoder(
		wavBuffer.Format.SampleRate,
		wavBuffer.Format.NumChannels,
	)
	return mp3Encoder.Write(of, decodedData)
}
