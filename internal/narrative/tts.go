package narrative

import (
	"path/filepath"
	"strconv"
)

type KittenModelType int

const (
	KittenModelMicro KittenModelType = iota
	KittenModelNano
	KittenModelMini
)

func GetKittenModel(t KittenModelType) *TTSModel {
	switch t {
	case KittenModelMicro:
		return &TTSModelMicro
	case KittenModelNano:
		return &TTSModelNano
	case KittenModelMini:
		return &TTSModelMini
	}
	panic("Unknown model type : " + strconv.Itoa(int(t)))
}

const (
	VOICES_FNAME      = "voices.npz"
	MODEL_LOCAL_FNAME = "kitten.onnx"
)

var (
	TTSModelNano = TTSModel{
		t:      KittenModelNano,
		name:   "nano",
		remote: "https://huggingface.co/KittenML/kitten-tts-nano-0.8/resolve/main",
		fname:  "kitten_tts_nano_v0_8.onnx",
		desc:   "(57 MB) Best for laptops and weaker PCs, offers good sound quality without straining CPU.",
	}
	TTSModelMicro = TTSModel{
		t:      KittenModelMicro,
		name:   "micro",
		remote: "https://huggingface.co/KittenML/kitten-tts-micro-0.8/resolve/main",
		fname:  "kitten_tts_micro_v0_8.onnx",
		desc:   "(41 MB) Smallest model available, suitable for older laptops and Raspberry Pi.",
	}
	TTSModelMini = TTSModel{
		t:      KittenModelMini,
		name:   "mini",
		remote: "https://huggingface.co/KittenML/kitten-tts-mini-0.8/resolve/main",
		fname:  "kitten_tts_mini_v0_8.onnx",
		desc:   "(78 MB) Best sound quality available, might strain CPU/GPU hence recommended for newer laptops and more powerful PCs.",
	}
)

// Represents model settings as saved in data config.
type TTSModel struct {
	t                         KittenModelType
	name, fname, remote, desc string
}

func (t TTSModel) FilterValue() string { return t.name }

func (t *TTSModel) GetModelDir(base string) string {
	return filepath.Join(base, t.name)
}

func (t *TTSModel) GetModelPath(base string) string {
	return filepath.Join(t.GetModelDir(base), MODEL_LOCAL_FNAME)
}

func (t *TTSModel) GetVoicesPath(base string) string {
	return filepath.Join(t.GetModelDir(base), VOICES_FNAME)
}
