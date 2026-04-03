package kitten

import (
	"fmt"
	"log"

	"github.com/sbinet/npyio/npz"
	ort "github.com/yalue/onnxruntime_go"
)

var VOICE_MAP map[string]string = map[string]string{
	"Bella":  "expr-voice-2-f.npy",
	"Jasper": "expr-voice-2-m.npy",
	"Luna":   "expr-voice-3-f.npy",
	"Bruno":  "expr-voice-3-m.npy",
	"Rosie":  "expr-voice-4-f.npy",
	"Hugo":   "expr-voice-4-m.npy",
	"Kiki":   "expr-voice-5-f.npy",
	"Leo":    "expr-voice-5-m.npy",
}

const DEFAULT_VOICE = "Luna"
const DEFAULT_SPEED float32 = 1.2

type vMat [400][256]float32

func (v *vMat) Load(flat []float32) {
	if len(flat) != 102400 {
		panic("Invalid voice matrix length, expected 102400")
	}
	for i := range 400 {
		v[i] = ([256]float32)(flat[i*256 : (i+1)*256])
	}

}

type Kitten struct {
	token_map *TokenMap
	voice     *vMat
	session   *ort.DynamicAdvancedSession
}

func (k *Kitten) Deinit() {
	k.session.Destroy()
	ort.DestroyEnvironment()
}

func (k *Kitten) createTensors(sentence []int64) (
	input *ort.Tensor[int64],
	voice *ort.Tensor[float32],
	speed *ort.Tensor[float32],
	err error,
) {
	sentence_l := len(sentence)
	sentence = append([]int64{0}, sentence...)
	sentence = append(sentence, 10)
	sentence = append(sentence, 0)

	input_tensor, err := ort.NewTensor(ort.NewShape(1, int64(len(sentence))), sentence)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating input tensor: %+v\n, %w", sentence, err)
	}

	var voice_i int = 399
	if sentence_l < 399 {
		voice_i = sentence_l
	}

	voice_data := k.voice[voice_i][:]

	voice_tensor, err := ort.NewTensor(ort.NewShape(1, 256), voice_data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating voice tensor: %w", err)
	}

	speed_tensor, err := ort.NewTensor(ort.NewShape(1), []float32{DEFAULT_SPEED})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error creating speed tensor: %w", err)
	}
	return input_tensor, voice_tensor, speed_tensor, nil
}

func (k *Kitten) RunInference(sentence string) ([]float32, error) {

	tokens := k.token_map.TokenizeWord(sentence)

	input_tensor, voice_tensor, speed_tensor, err := k.createTensors(tokens)
	if err != nil {
		return nil, err
	}
	defer input_tensor.Destroy()
	defer voice_tensor.Destroy()
	defer speed_tensor.Destroy()

	outputs := []ort.Value{nil, nil}
	defer func() {
		for _, o := range outputs {
			o.Destroy()
		}
	}()
	err = k.session.Run(
		[]ort.Value{input_tensor, voice_tensor, speed_tensor},
		outputs,
	)
	if err != nil {
		return nil, fmt.Errorf("Error running inference: %w", err)
	}
	waveform := outputs[0].(*ort.Tensor[float32]).GetData()
	return waveform, nil
}

func NewKitten(voice_name *string) *Kitten {
	ort.SetSharedLibraryPath("/home/exaroth/Projects/narrative/lib/libonnxruntime.so")

	err := ort.InitializeEnvironment()
	if err != nil {
		log.Fatalf("Could intialize onnx runtime: %+v", err)
	}
	var voice_dtf string
	if voice_name != nil {
		v, ok := VOICE_MAP[*voice_name]
		if !ok {
			log.Fatalf("Invalid voice id %s provided", *voice_name)
		}
		voice_dtf = v
	} else {
		voice_dtf = VOICE_MAP[DEFAULT_VOICE]
	}

	f, err := npz.Open("./models/kitten/voices.npz")
	if err != nil {
		log.Fatalf("Could not open npz file: %+v", err)
	}
	defer f.Close()

	var f0 []float32

	err = f.Read(voice_dtf, &f0)
	if err != nil {
		log.Fatalf("Could not read value from npz file: %+v", err)
	}
	var voice vMat
	voice.Load(f0)

	session, err := ort.NewDynamicAdvancedSession(
		"./models/kitten/kitten.onnx",
		[]string{"input_ids", "style", "speed"},
		[]string{"waveform", "duration"},
		nil,
	)
	if err != nil {
		log.Fatalf("Error intializing session %+v", err)
	}

	token_map := BuildTokenMap()

	return &Kitten{
		token_map: &token_map,
		voice:     &voice,
		session:   session,
	}

}
