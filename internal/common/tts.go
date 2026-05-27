package common

import (
	"fmt"
	"path/filepath"
	"strconv"
)

type KittenModelType int

const (
	KittenModelMicro KittenModelType = iota
	KittenModelNano
	KittenModelMini
)

func (k KittenModelType) String() string {
	switch k {
	case KittenModelMicro:
		return "micro"
	case KittenModelNano:
		return "nano"
	case KittenModelMini:
		return "mini"
	default:
		panic("Unknown model " + strconv.Itoa(int(k)))
	}
}

// Retrieve kitten type from string
func KittenTypeFromString(n string) (KittenModelType, error) {
	switch n {
	case "nano":
		return KittenModelNano, nil
	case "micro":
		return KittenModelMicro, nil
	case "mini":
		return KittenModelMini, nil
	default:
		return -1, fmt.Errorf("Unknown model")
	}
}

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
		T:      KittenModelNano,
		Name:   "nano",
		Remote: "https://huggingface.co/KittenML/kitten-tts-nano-0.8/resolve/main",
		Fname:  "kitten_tts_nano_v0_8.onnx",
		Desc:   "(57 MB) Best for laptops and weaker PCs, offers good sound quality without straining CPU.",
	}
	TTSModelMicro = TTSModel{
		T:      KittenModelMicro,
		Name:   "micro",
		Remote: "https://huggingface.co/KittenML/kitten-tts-micro-0.8/resolve/main",
		Fname:  "kitten_tts_micro_v0_8.onnx",
		Desc:   "(41 MB) Smallest model, suitable for older laptops and Raspberry Pi, not recommended otherwise.",
	}
	TTSModelMini = TTSModel{
		T:      KittenModelMini,
		Name:   "mini",
		Remote: "https://huggingface.co/KittenML/kitten-tts-mini-0.8/resolve/main",
		Fname:  "kitten_tts_mini_v0_8.onnx",
		Desc:   "(78 MB) Best sound quality available, might strain CPU/GPU hence recommended for newer laptops and more powerful PCs.",
	}
)

// Represents model settings as saved in data config.
type TTSModel struct {
	T                         KittenModelType
	Name, Fname, Remote, Desc string
}

func (t TTSModel) FilterValue() string { return t.Name }

func (t *TTSModel) GetModelDir(base string) string {
	return filepath.Join(base, t.Name)
}

func (t *TTSModel) GetModelPath(base string) string {
	return filepath.Join(t.GetModelDir(base), MODEL_LOCAL_FNAME)
}

func (t *TTSModel) GetVoicesPath(base string) string {
	return filepath.Join(t.GetModelDir(base), VOICES_FNAME)
}
