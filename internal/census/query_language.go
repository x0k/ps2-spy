package census

type censusLanguage int

const (
	LangEnglish censusLanguage = iota
	LangGerman
	LangSpanish
	LangFrench
	LangItalian
	LangTurkish
)

var censusLanguages = []string{"en", "de", "es", "fr", "it", "tr"}
