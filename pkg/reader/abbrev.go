package reader

import (
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	ABBREV_CASE_I = `(\s|^)(adj|asst|ave|bld|bldg|blvd|bros|btw|capt|cmdr|co|col|conn|corp|cpl|` +
		`dec|dept|dr|drs|e[.]g|eg|feb|ft|esq|gov|hon|hosp|hr|hrs|hway|jun|` +
		`gen|hwy|i[.]e|ie|inc|insp|jan|jr|jul|lt|ltd|maj|mar|mass|md|nov|` +
		`max|min|mr|mrs|ms|msgr|messrs|mmes|mses|miss|nebr|nev|nos|apr|` +
		`nr|oct|ok|ph[.]d|phd|ny|penn|pls|prof|n[.]y|pvt|ref|rev|rep|sec|sep|` +
		`sept|sgt|sr|st|tenn|tex|univ|us|u[.]s|ver|vs|fig|brig|att|sen|adm|aug` +
		`)\b([.])`
	ABBREV_CASE_I_RE = regexp.MustCompile(`(?i)` + ABBREV_CASE_I)

	ABBREV_CASE_S_RE = regexp.MustCompile(`\s(c[.]|s[.]|p[.]|v[.]|no[.])(\s|$)`)
)

var ABBREV_MAP = map[string]string{
	"adj": "adjective", "asst": "assistant", "ave": "avenue", "bld": "building", "bldg": "building",
	"blvd": "boulevard", "bros": "brothers", "btw": "by the way", "capt": "captain", "cmdr": "commander",
	"co": "company", "col": "colonel", "conn": "connection", "corp": "corporation", "cpl": "corporal",
	"dec": "December", "dept": "department", "dr": "doctor", "drs": "doctor's", "e.g": "for example",
	"eg": "for example", "feb": "February", "ft": "featuring", "esq": "esquire", "gov": "government",
	"hon": "honorable", "hosp": "hospital", "hr": "hour", "hrs": "hours", "hway": "highway",
	"gen": "general", "hwy": "highway", "i.e": "that is", "ie": "that is", "inc": "incorporated",
	"insp": "inspector", "jan": "January", "jr": "junior", "Jul": "July", "lt": "lieutenant",
	"ltd": "limited", "maj": "major", "mar": "March", "mass": "massachusetts",
	"md": "medical doctor", "max": "maximum", "min": "minimum", "mr": "mister", "mrs": "missus",
	"ms": "miss", "msgr": "monsignor", "messrs": "misters", "mmes": "mesdames", "mses": "misses",
	"miss": "miss", "nebr": "nebraska", "nev": "nevada", "no.": "number", "nos": "numbers",
	"nr": "number", "oct": "October", "ok": "ok", "ph.d": "doctor of physics", "aug": "August",
	"phd": "doctor of physics", "ny": "new york", "penn": "pennsylvania", "pls": "please",
	"prof": "professor", "n.y": "new york", "pvt": "private", "ref": "reference", "nov": "November",
	"rev": "reverend", "rep": "representative", "sec": "second", "sep": "September", "sept": "September",
	"sgt": "sergeant", "sr": "senior", "st": "street", "tenn": "tennessee", "tex": "texas", "univ": "university",
	"us": "united states", "u.s": "united states", "ver": "version", "vs": "versus", "fig": "figure",
	"brig": "brigadier", "att": "attorney", "sen": "senator", "adm": "admiral", "apr": "April", "jun": "June",
	"c.": "cent", "s.": "shilling", "p.": "pence", "v.": "version",
}

// Expand common abbreviations in the input string, this is done early
// as leaving those in might intefere with properly splitting sentences.
func ExpandAbbreviations(in string) string {
	var key string
	for _, m := range ABBREV_CASE_I_RE.FindAllStringIndex(in, 1) {
		key = strings.ToLower(strings.Trim(in[m[0]:m[1]], ". "))
		if abb, ok := ABBREV_MAP[key]; ok {
			ns := in[:m[0]] + " " + abb + " " + in[m[1]:]
			return ExpandAbbreviations(ns)
		} else {
			logrus.Warning("FIXME: Missing abbreviation " + key)
		}
	}
	for _, m := range ABBREV_CASE_S_RE.FindAllStringIndex(in, 1) {
		key = strings.Trim(in[m[0]:m[1]], " ")
		if abb, ok := ABBREV_MAP[key]; ok {
			ns := in[:m[0]] + " " + abb + " " + in[m[1]:]
			return ExpandAbbreviations(ns)
		} else {
			logrus.Warning("FIXME: Missing abbreviation " + key)
		}
	}
	return in
}
