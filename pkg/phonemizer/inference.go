package phonemizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/neurlang/classifier/datasets/phonemizer_ulevel"
	"github.com/neurlang/classifier/hash"
	"github.com/neurlang/classifier/layer/crossattention"
	"github.com/neurlang/classifier/layer/sochastic"
	"github.com/neurlang/classifier/layer/sum"
	"github.com/neurlang/classifier/net/feedforward"

	dict "github.com/exaroth/narrative/dictionary"
)

func hashtronHash(str string) uint32 {
	return hash.StringHash(0, str)
}

type HashtronPhonemizer struct {
	reverse   bool
	dict_path *string
	mut       *sync.RWMutex
	lang      *language
	network   *feedforward.FeedforwardNetwork
}

func NewHashtronPhonemizer(dict_path *string, reverse bool) *HashtronPhonemizer {

	return &HashtronPhonemizer{
		reverse:   reverse,
		dict_path: dict_path,
		lang:      nil,
		network:   nil,
		mut:       &sync.RWMutex{},
	}
}

func (r *HashtronPhonemizer) LoadLanguage() error {
	var reverse string
	if r.reverse {
		reverse = "_reverse"
	}
	// Double-checked locking: acquire write lock and check again
	r.mut.Lock()
	defer r.mut.Unlock()

	var language_file = "language" + reverse + ".json"

	// log.Now().Debugf("Language %s loading file", file)
	f_contents, err := dict.Language.ReadFile(language_file)
	if err != nil {
		return err
	}

	// Parse the JSON data into the Language struct
	var langone language
	err = json.Unmarshal(f_contents, &langone)
	if err != nil {
		return fmt.Errorf("Error parsing JSON: %v\n", err)
	}

	langone.mapize()
	langone.srcdst()
	langone.letters()

	r.lang = &langone

	var weights_file = "weights6" + reverse + ".json.zlib"

	f_contents, err = dict.Language.ReadFile(weights_file)
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

	r.network = &net
	err = r.network.ReadZlibWeights(bytesReader)

	return err
}

func (r *HashtronPhonemizer) PhonemizeWord(word string) (ret []map[string]uint32, err error) {

	var backoffs = 10

	srca := strings.Split(word, "")
	dsta := []string{}
	var lastspace = 0

outer:
	for i := 0; i < len(srca); i++ {
		srcv := srca[i]
		m := r.lang.Mapping[srcv]

		if len(m) == 0 {
			dsta = append(dsta, "")
			continue
		}

		if len(m) == 1 {
			for _, mfirst := range m {
				if mfirst == "_" {
					lastspace = i + 1
				} else if strings.HasPrefix(mfirst, "_") {
					lastspace = i
				} else if strings.HasSuffix(mfirst, "_") {
					lastspace = i + 1
				}
				dsta = append(dsta, mfirst)
				break
			}
			continue
		}

		for _, option := range m {
			srcaR := srca[lastspace:]
			dstaR := dsta[lastspace:]
			origi := i
			i := i - lastspace

			var multiword = lastspace > 0
			var predicted int

			for q := 0; (!multiword && q == 0) || (multiword && q < len(srcaR)-i); q++ {

				const fanout1new = 24
				var input2 = phonemizer_ulevel.NewInferenceSubsample(srcaR, dstaR, option, fanout1new/3)
				var pred int

				r.mut.RLock()
				pred = int(r.network.Infer2(input2))
				r.mut.RUnlock()

				predicted += pred

				// fmt.Printf("Model predicted: %v %v %v -> %d\n", srcaR, dstaR, option, pred)
			}

			if (!multiword && predicted == 1) || (multiword && 2*predicted > len(srcaR)) {
				if option == "_" {
					lastspace = origi + 1
				} else if strings.HasPrefix(option, "_") {
					lastspace = origi
				} else if strings.HasSuffix(option, "_") {
					lastspace = origi + 1
				}
				dsta = append(dsta, option)
				continue outer
			}
		}
		if backoffs > 0 {
			i = lastspace - 1
			dsta = dsta[:lastspace]
			backoffs--
			continue
		}
		for _, mfirst := range m {
			if mfirst == "_" {
				lastspace = i + 1
			} else if strings.HasPrefix(mfirst, "_") {
				lastspace = i
			} else if strings.HasSuffix(mfirst, "_") {
				lastspace = i + 1
			}
			dsta = append(dsta, mfirst)
			break
		}
	}
	var src, dst string

	push := func() {
		if len(src)+len(dst) > 0 {
			m := make(map[string]uint32)
			hsh := hashtronHash(src + "\x00" + dst)
			if hsh == 0 {
				hsh++
			}
			m[dst] = uint32(hsh)
			m[src+" "] = 0
			ret = append(ret, m)
			src, dst = "", ""
		}
	}
	for i, v := range dsta {
		if v != "_" && strings.HasPrefix(v, "_") {
			push()
		}
		src += srca[i]
		dst += v
		if strings.HasSuffix(v, "_") {
			push()
		}
	}
	push()
	return
}
