package phonemizer

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	dict "github.com/exaroth/narrative/dictionary"
	"github.com/neurlang/classifier/hash"
	"github.com/neurlang/classifier/layer/crossattention"
	"github.com/neurlang/classifier/layer/sochastic"
	"github.com/neurlang/classifier/layer/sum"
	"github.com/neurlang/classifier/net/feedforward"
)

const HOMONYM_WEIGHTS_FNAME = "weights7.json.zlib"

// Controller for handling phoneme selection,
// uses goruut weights file for inference.
type PhonemeSelector struct {
	network *feedforward.FeedforwardNetwork
	mut     *sync.RWMutex
}

// Select preferred phonemes for given sentence.
func (h *PhonemeSelector) Select(sentence []map[string][2]uint32) (ret [][3]uint32) {

	var ai_sentence = PhonemizerSample{
		Sentence: []PhonemizerToken{},
	}

	for _, mapping := range sentence {
		var origword string
		var strkey [][3]string
		for v, k := range mapping {
			if k[0] == 0 {
				origword = strings.TrimRight(v, " ")
				continue
			}
			strkey = append(strkey, [3]string{v, fmt.Sprint(k[0]), fmt.Sprint(k[1])})
		}
		sort.SliceStable(strkey, func(i, j int) bool {
			return strkey[i][0] < strkey[j][0]
		})
		var choices [][2]uint32
		for _, v := range strkey {
			num, _ := strconv.Atoi(v[1])
			improvised, _ := strconv.Atoi(v[2])
			if improvised == 1 {
				choices = append(choices, [2]uint32{0, uint32(num)})
			} else {
				choices = append(choices, [2]uint32{hash.StringHash(0, v[0]), uint32(num)})
			}
		}
		sort.SliceStable(choices, func(i, j int) bool {
			return choices[i][0] < choices[j][0]
		})
		var sol uint32
		if len(choices) > 0 {
			sol = choices[0][0]
		}

		ai_sentence.Sentence = append(ai_sentence.Sentence, PhonemizerToken{
			Homograph: hash.StringHash(0, origword),
			Choices:   choices,
			Solution:  sol,
		})
	}
	const fanout1 = 24
	for i := range ai_sentence.Sentence {
		var sample = ai_sentence.V2(fanout1, i)
		if sample.Len() <= 1 {
			// no choice
			continue
		}
		var unchosen, chosen [2]uint32
		var accept bool
		for j := 0; !accept && j < sample.Len(); j++ {
			ai_sentence.Sentence[i].Solution = ai_sentence.Sentence[i].Choices[j][0]
			var pred uint32

			h.mut.RLock()
			pred = uint32(h.network.Infer2(sample.IO(j)))
			h.mut.RUnlock()

			if pred == 1 && !accept {
				accept = true
				chosen = ai_sentence.Sentence[i].Choices[j]
			} else if j == 0 {
				unchosen = ai_sentence.Sentence[i].Choices[j]
			}
		}
		var pred uint32
		if !accept {
			ai_sentence.Sentence[i].Solution = unchosen[0]
			pred = unchosen[1]
		} else {
			ai_sentence.Sentence[i].Solution = chosen[0]
			pred = chosen[1]
		}
		ret = append(ret, [3]uint32{uint32(i), ai_sentence.Sentence[i].Solution, pred})
	}
	return

}

// Load weights file and initialize inference network.
func (h *PhonemeSelector) LoadLanguage() error {

	f_contents, err := dict.Language.ReadFile(HOMONYM_WEIGHTS_FNAME)
	if err != nil {
		return err
	}

	bytesReader := bytes.NewReader(f_contents)

	const fanout1 = 24
	const fanout2 = 1
	const fanout3 = 4
	const fanout4 = 32

	var net feedforward.FeedforwardNetwork
	net.NewLayer(fanout1*fanout2, 0)
	for i := range fanout3 {
		net.NewCombiner(crossattention.MustNew3(fanout1, fanout2))
		net.NewLayerPI(fanout1*fanout2, 0, 0)
		net.NewCombiner(sochastic.MustNew(fanout1*fanout2, fanout4-8*byte(i), uint32(i)))
		net.NewLayerPI(fanout1*fanout2, 0, 0)
	}
	net.NewCombiner(sochastic.MustNew(fanout1*fanout2, fanout4, fanout3))
	net.NewLayer(fanout1*fanout2, 0)
	net.NewCombiner(sum.MustNew([]uint{fanout1 * fanout2}, 0))
	net.NewLayer(1, 0)

	h.network = &net

	return h.network.ReadZlibWeights(bytesReader)
}

// Initialize new phoneme selector.
func NewPhonemeSelector() *PhonemeSelector {

	return &PhonemeSelector{
		mut:     &sync.RWMutex{},
		network: nil,
	}
}
