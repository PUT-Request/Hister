package document

import (
	"strings"

	// register lingua-go language packs
	"github.com/asciimoo/lingua-go"
	_ "github.com/asciimoo/lingua-go/language-models/ar"
	_ "github.com/asciimoo/lingua-go/language-models/bg"
	_ "github.com/asciimoo/lingua-go/language-models/ca"
	_ "github.com/asciimoo/lingua-go/language-models/da"
	_ "github.com/asciimoo/lingua-go/language-models/de"
	_ "github.com/asciimoo/lingua-go/language-models/el"
	_ "github.com/asciimoo/lingua-go/language-models/en"
	_ "github.com/asciimoo/lingua-go/language-models/es"
	_ "github.com/asciimoo/lingua-go/language-models/eu"
	_ "github.com/asciimoo/lingua-go/language-models/fa"
	_ "github.com/asciimoo/lingua-go/language-models/fi"
	_ "github.com/asciimoo/lingua-go/language-models/fr"
	_ "github.com/asciimoo/lingua-go/language-models/ga"
	_ "github.com/asciimoo/lingua-go/language-models/hi"
	_ "github.com/asciimoo/lingua-go/language-models/hr"
	_ "github.com/asciimoo/lingua-go/language-models/hu"
	_ "github.com/asciimoo/lingua-go/language-models/hy"
	_ "github.com/asciimoo/lingua-go/language-models/id"
	_ "github.com/asciimoo/lingua-go/language-models/it"
	_ "github.com/asciimoo/lingua-go/language-models/ja"
	_ "github.com/asciimoo/lingua-go/language-models/ko"
	_ "github.com/asciimoo/lingua-go/language-models/nb"
	_ "github.com/asciimoo/lingua-go/language-models/nl"
	_ "github.com/asciimoo/lingua-go/language-models/pl"
	_ "github.com/asciimoo/lingua-go/language-models/pt"
	_ "github.com/asciimoo/lingua-go/language-models/ro"
	_ "github.com/asciimoo/lingua-go/language-models/ru"
	_ "github.com/asciimoo/lingua-go/language-models/sv"
	_ "github.com/asciimoo/lingua-go/language-models/tr"
	_ "github.com/asciimoo/lingua-go/language-models/zh"
)

const UnknownLanguage = "unknown"

var Languages = []lingua.Language{
	lingua.Arabic,    // ar
	lingua.Bokmal,    // nb - Norewgian Bokmal, gets rewritten to "no" in Hister
	lingua.Bulgarian, // bg
	lingua.Catalan,   // ca
	lingua.Chinese,   // zh, uses the CJK analyzer
	// lingua.Czech,      // cs
	lingua.Danish,     // da
	lingua.German,     // de
	lingua.Greek,      // el
	lingua.English,    // en
	lingua.Spanish,    // es
	lingua.Basque,     // eu
	lingua.Persian,    // fa
	lingua.Finnish,    // fi
	lingua.French,     // fr
	lingua.Irish,      // ga
	lingua.Hindi,      // hi
	lingua.Croatian,   // hr
	lingua.Hungarian,  // hu
	lingua.Armenian,   // hy
	lingua.Indonesian, // id
	lingua.Italian,    // it
	lingua.Japanese,   // ja, uses the CJK analyzer
	lingua.Korean,     // ko, uses the CJK analyzer
	lingua.Dutch,      // nl
	lingua.Polish,     // pl
	lingua.Portuguese, // pt
	lingua.Romanian,   // ro
	lingua.Russian,    // ru
	lingua.Swedish,    // sv
	lingua.Turkish,    // tr
	// supported by bleve but not by lingua: gl, in
}

// LanguageDetector detects the language of a text.
type LanguageDetector interface {
	DetectLanguage(string) string
}

type nullLangDetector struct{}

type langDetector struct {
	detector lingua.LanguageDetector
}

func NewLanguageDetector() LanguageDetector {
	return &langDetector{
		detector: lingua.NewLanguageDetectorBuilder().FromLanguages(Languages...).Build(),
	}
}

func NewNullLanguageDetector() LanguageDetector {
	return &nullLangDetector{}
}

func (d *langDetector) DetectLanguage(s string) string {
	if language, exists := d.detector.DetectLanguageOf(s); exists {
		code := strings.ToLower(language.IsoCode639_1().String())
		switch code {
		case "nb":
			// use generic "no" code for Norwegian
			return "no"
		}
		return code
	}
	return UnknownLanguage
}

func (d *nullLangDetector) DetectLanguage(s string) string {
	return UnknownLanguage
}
