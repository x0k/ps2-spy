package census

type CensusLanguage int

const (
	LangEnglish CensusLanguage = iota
	LangGerman
	LangSpanish
	LangFrench
	LangItalian
	LangTurkish
)

var censusLanguages = []string{"en", "de", "es", "fr", "it", "tr"}
