package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/exaroth/narrative/phonemizer"
	"github.com/sbinet/npyio/npz"
	log "github.com/sirupsen/logrus"
	ort "github.com/yalue/onnxruntime_go"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

const (
	pad = '$'
)

var chars_punctuation = []rune{';', ':', ',', '.', '!', '?', '¡', '¿', '—', '…', '"', '«', '»', '"', '"', ' '}
var chars_letter = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var chars_ipa = []rune("ɑɐɒæɓʙβɔɕçɗɖðʤəɘɚɛɜɝɞɟʄɡɠɢʛɦɧħɥʜɨɪʝɭɬɫɮʟɱɯɰŋɳɲɴøɵɸθœɶʘɹɺɾɻʀʁɽʂʃʈʧʉʊʋⱱʌɣɤʍχʎʏʑʐʒʔʡʕʢǀǁǂǃˈˌːˑʼʴʰʱʲʷˠˤ˞↓↑→↗↘'̩'ᵻ")

const testText = "When I first tackled this, I made the mistake of not aligning the library versions, which caused unexpected runtime errors. So let’s get it right from the start."

type vMat [400][256]float32

func (v *vMat) Load(flat []float32) {
	if len(flat) != 102400 {
		panic("Invalid voice matrix length, expected 102400")
	}
	for i := range 400 {
		v[i] = ([256]float32)(flat[i*256 : (i+1)*256])
	}

}

type TokenMap map[rune]int64

func (t TokenMap) tokenize(char rune) int64 {
	if val, ok := t[char]; ok {
		return val
	}
	fmt.Printf("char %s not found in token map", string(char))
	return -1
}

func (t TokenMap) TokenizeWord(word string) []int64 {
	result := []int64{}
	for _, char := range word {
		result = append(result, t.tokenize(char))
	}
	return result
}

func buildTokenMap() TokenMap {
	token_arr := []rune{pad}
	token_arr = append(token_arr, chars_punctuation...)
	token_arr = append(token_arr, chars_letter...)
	token_arr = append(token_arr, chars_ipa...)

	var result = make(map[rune]int64)
	for idx, c := range token_arr {
		result[c] = int64(idx)
	}
	return result
}

func main() {

	token_map := buildTokenMap()

	fmt.Println(token_map.TokenizeWord("1"))

	phonemizer, err := phonemizer.NewPhonemizer()
	if err != nil {
		panic(err)
	}
	r, _ := phonemizer.Phonemize("deprofundis")
	fmt.Println(r)

	f, err := npz.Open("./kitten/voices.npz")
	if err != nil {
		log.Fatalf("could not open npz file: %+v", err)
	}
	defer f.Close()

	// for _, name := range f.Keys() {
	// 	fmt.Println(name)
	// 	fmt.Printf("%s: %v\n", name, f.Header(name))
	// 	fmt.Println(f.Header(name).Descr)
	// }

	var f0 []float32

	err = f.Read("expr-voice-2-m.npy", &f0)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}
	var out vMat
	out.Load(f0)
	// fmt.Println(out)

	ort.SetSharedLibraryPath("/home/exaroth/Projects/narrative/onnx_libs/lib/libonnxruntime.so")

	err = ort.InitializeEnvironment()
	if err != nil {
		panic(err)
	}
	defer ort.DestroyEnvironment()

	inputData := []int64{0, 46, 47, 58, 60, 57, 48, 63, 56, 46, 51, 61, 16, 45, 54, 43, 55, 57, 16, 43, 46, 16, 62, 47, 16, 46, 57, 55, 51, 56, 47, 10, 0}
	inputShape := ort.NewShape(1, 24)
	inputTensor, err := ort.NewTensor(inputShape, inputData)
	defer inputTensor.Destroy()
	voiceData := out[24][:]
	// fmt.Println(voiceData)

	voiceShape := ort.NewShape(1, 256)
	voiceTensor, err := ort.NewTensor(voiceShape, voiceData)
	defer voiceTensor.Destroy()
	speedShape := ort.NewShape(1)
	speedTensor, err := ort.NewTensor(speedShape, []float32{1.0})
	// This hypothetical network maps a 2x5 input -> 2x3x4 output.

	session, err := ort.NewDynamicAdvancedSession(
		"./kitten/kitten.onnx",
		[]string{"input_ids", "style", "speed"},
		[]string{"waveform"},
		nil,
	)

	// defer session.Destroy()

	outputShape := ort.NewShape(78000)
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	defer outputTensor.Destroy()

	// // duration, err := ort.NewEmptyScalar[int64]()
	// duration, err := ort.NewEmptyTensor[int64](ort.NewShape())
	// defer duration.Destroy()

	err = session.Run(
		[]ort.Value{inputTensor, voiceTensor, speedTensor},
		[]ort.Value{outputTensor},
	)
	fmt.Println("err")
	fmt.Println(err)

	// Get a slice view of the output tensor's data.
	outputData := outputTensor.GetData()
	// fmt.Println("out")
	// fmt.Println(outputData)

	fname := "out.bin"
	file, _ := os.Create(fname)
	// decayfac := math.Pow(end/start, 1.0/float64(nsamps))
	for _, sample := range outputData {
		// fmt.Println(sample)
		// fmt.Println(idx)
		// sample := math.Sin(angle * Frequency * float64(i))
		// sample *= start
		// start *= decayfac
		var buf [8]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(float32(sample)))
		_, err := file.Write(buf[:])
		if err != nil {
			panic(err)
		}
		// fmt.Printf("\rWrote: %v bytes to %s", bw, file)
	}
}
