package phonemizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode"

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
	// phoner  *interfaces.Phonemizer
	network *feedforward.FeedforwardNetwork
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
	for i := 0; i < fanout3; i++ {
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

type language struct {
	Mapping        map[string][]string `json:"Map"`
	SrcMulti       []string            `json:"SrcMulti"`
	DstMulti       []string            `json:"DstMulti"`
	SrcMultiSuffix []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix []string            `json:"DstMultiSuffix"`
	DropLast       []string            `json:"DropLast"`
	SrcDuplicate   [][]string          `json:"SrcDuplicate"`
	//Histogram         []string            `json:"Histogram"`
	mapTokenizer      map[uint32]map[[2]uint32]string
	mapSrcMultiLen    int
	mapSrcMultiSufLen int
	mapSrcMulti       map[string]struct{}
	mapDstMulti       map[string]struct{}
	mapSrcMultiSuffix map[string]struct{}
	mapDstMultiSuffix map[string]struct{}
	mapLetters        map[string]struct{}
	mapDropLast       map[string]struct{}
}

func mapize(arr []string) (out map[string]struct{}) {
	out = make(map[string]struct{})
	for _, v := range arr {
		out[v] = struct{}{}
	}
	return
}

func (l *language) mapize() {
	// todo
	// l.mapTokenizer = noareg.MakeDetokenizer(l.Mapping)
	l.mapSrcMulti = mapize(l.SrcMulti)
	l.mapDstMulti = mapize(l.DstMulti)
	l.mapSrcMultiSuffix = mapize(l.SrcMultiSuffix)
	l.mapDstMultiSuffix = mapize(l.DstMultiSuffix)
	l.mapDropLast = mapize(l.DropLast)
	l.SrcMulti = nil
	l.DstMulti = nil
	l.SrcMultiSuffix = nil
	l.DstMultiSuffix = nil
	l.DropLast = nil
}

func (l *language) srcdst() {
	for k, v := range l.Mapping {
		if v == nil || len(v) == 0 {
			continue
		}
		if len([]rune(k)) > 1 {
			l.mapSrcMulti[k] = struct{}{}
		}
		for _, w := range v {
			if len([]rune(w)) > 1 {
				l.mapDstMulti[w] = struct{}{}
			}
		}
	}
	for k := range l.mapSrcMulti {
		if len([]rune(k)) > l.mapSrcMultiLen {
			l.mapSrcMultiLen = len([]rune(k))
		}
	}
	for k := range l.mapSrcMultiSuffix {
		if len([]rune(k)) > l.mapSrcMultiSufLen {
			l.mapSrcMultiSufLen = len([]rune(k))
		}
	}
}

func (l *language) letters() {
	l.mapLetters = make(map[string]struct{})
	for k := range l.Mapping {
		addLetters(k, l.mapLetters)
	}
	for _, rule := range l.SrcDuplicate {
		for _, v := range rule {
			addLetters(v, l.mapLetters)
		}
	}
}

func isCombining(r uint32) bool {
	return unicode.Is(unicode.Mn, rune(r)) || unicode.Is(unicode.Mc, rune(r))
}

func addLetters(word string, mapping map[string]struct{}) {
	if mapping == nil {
		return
	}

	runes := []rune(word)
	n := len(runes)
	var str string
	var baseLetterFound bool

	// Process characters in reverse order
	for i := n - 1; i >= 0; i-- {
		r := runes[i]

		if isCombining(uint32(r)) {
			// Accumulate combining characters
			str = string(r) + str
		} else {
			// Found a base letter, prepend accumulated combiners
			fullLetter := string(r) + str
			// log.Now().Debugf("Adding to letters: %s", fullLetter)
			mapping[fullLetter] = struct{}{}

			// Reset for next letter
			str = ""
			baseLetterFound = true
		}
	}

	// Edge case: if only combiners were present at the start
	if str != "" && !baseLetterFound {
		// log.Now().Debugf("Adding standalone combiners: %s", str)
		mapping[str] = struct{}{}
	}
}
