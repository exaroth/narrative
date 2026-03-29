package phonemizer

type PhonemizerToken struct {
	// homograph = hash of written word == query
	Homograph uint32
	// solution = hash of ipa word == value
	Solution uint32
	// here the fisrt integer is like solution (hash of ipa word), the second is the tag key
	Choices [][2]uint32
}

type PhonemizerSample struct {
	Sentence []PhonemizerToken
}

func (t *PhonemizerToken) Len() int {
	return len(t.Choices)
}

func (s *PhonemizerSample) V1(dim, pos int) SampleSentence {
	return SampleSentence{
		Sample:    s,
		position:  pos,
		dimension: dim,
		version:   1,
	}
}
func (s *PhonemizerSample) V2(dim, pos int) SampleSentence {
	return SampleSentence{
		Sample:    s,
		position:  pos,
		dimension: dim,
		version:   2,
	}
}

type SampleSentence struct {
	Sample    *PhonemizerSample
	position  int
	dimension int
	version   byte
}

func (s *SampleSentence) Len() int {
	if len(s.Sample.Sentence) > s.position {
		return s.Sample.Sentence[s.position].Len()
	}
	return 0
}

type SampleSentenceIO struct {
	SampleSentence *SampleSentence
	choice         int
}

func (s *SampleSentence) IO(n int) (ret *SampleSentenceIO) {
	return &SampleSentenceIO{
		SampleSentence: s,
		choice:         n,
	}
}

// Feature: calculates query, key, value input for attention matrix
// n - if dividible by 3, it's supposed to return the homograph
// n - if equal to 1 divided by 3, it calculates the key token
// n - if equal to 2 divided by 3, it calculates the value token
func (s *SampleSentenceIO) Feature(n int) (ret uint32) {
	pos := (n / 3) % (s.SampleSentence.dimension / 3)
	if s.Parity() == 1 {
		ret = 1 << 31
	}
	if n%3 == 0 {
		for ; pos < len((s.SampleSentence.Sample.Sentence)); pos += (s.SampleSentence.dimension / 3) {
			ret += uint32(s.SampleSentence.Sample.Sentence[pos].Homograph)
		}
		return

	}
	for ; pos < len((s.SampleSentence.Sample.Sentence)); pos += (s.SampleSentence.dimension / 3) {
		if pos < s.SampleSentence.position {
			ret += uint32(s.SampleSentence.Sample.Sentence[pos].Solution)
		} else if pos == s.SampleSentence.position {
			choice := s.SampleSentence.Sample.Sentence[pos].Choices[s.choice]
			// Compare current choice with context
			if n%3 == 1 {
				ret += uint32(choice[1]) // Key
			} else if n%3 == 2 {
				ret += uint32(choice[0]) // Value
			}
		} else if s.SampleSentence.version >= 2 {
			for _, choice := range s.SampleSentence.Sample.Sentence[pos].Choices {
				// Compare future shifted choice with context
				if n%3 == 1 {
					ret += uint32(choice[1]) >> 16 // Key
				} else if n%3 == 2 {
					ret += uint32(choice[0]) >> 16 // Value
				}
			}
		}
	}
	return
}

func (s *SampleSentenceIO) Parity() (ret uint16) {
	return uint16(len(s.SampleSentence.Sample.Sentence) & 1)
	//return 0
}
func (s *SampleSentenceIO) Output() (ret uint16) {
	if s.SampleSentence.Sample.Sentence[s.SampleSentence.position].Choices[s.choice][0] == s.SampleSentence.Sample.Sentence[s.SampleSentence.position].Solution {
		return 1
	}
	return 0
}
