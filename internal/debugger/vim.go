package debugger

import (
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
)

type SelectSentenceCmd int

func SelectSentence(val int) tea.Cmd {
	return func() tea.Msg {
		return SelectSentenceCmd(val)
	}
}

type SelectWordCmd int

func SelectWord(val int) tea.Cmd {
	return func() tea.Msg {
		return SelectWordCmd(val)
	}
}

type VimPassthrough struct {
	value   int
	valueS  string
	start   time.Time
	lastUpd time.Time
}

func NewVimPassthrough(initial string) *VimPassthrough {
	i, _ := strconv.Atoi(initial)
	return &VimPassthrough{
		start:   time.Now(),
		lastUpd: time.Now(),
		valueS:  initial,
		value:   i,
	}
}

func (v *VimPassthrough) Update(key string) tea.Cmd {
	_, err := strconv.Atoi(key)
	if err == nil {
		v.valueS += key
		v.value, _ = strconv.Atoi(v.valueS)
		v.lastUpd = time.Now()
		return nil
	} else {
		if key == "w" {
			return SelectWord(v.value - 1)
		}
		if key == "s" {
			return SelectSentence(v.value - 1)
		}

	}
	return nil
}

func (f *VimPassthrough) Tick(t time.Time) bool {
	r := int(t.Sub(f.lastUpd).Seconds()) > 1

	return r
}
