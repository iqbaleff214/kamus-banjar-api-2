package domain

// WordClass represents the grammatical category of a word.
type WordClass string

const (
	WordClassNomina    WordClass = "n"
	WordClassVerba     WordClass = "v"
	WordClassAdjektiva WordClass = "a"
	WordClassAdverbia  WordClass = "adv"
	WordClassPartikel  WordClass = "p"
	WordClassPribahasa WordClass = "pb"
	WordClassKiasan    WordClass = "ki"
	WordClassNumeralia WordClass = "num"
	WordClassPronomina WordClass = "pron"
)

var validWordClasses = map[WordClass]struct{}{
	WordClassNomina:    {},
	WordClassVerba:     {},
	WordClassAdjektiva: {},
	WordClassAdverbia:  {},
	WordClassPartikel:  {},
	WordClassPribahasa: {},
	WordClassKiasan:    {},
	WordClassNumeralia: {},
	WordClassPronomina: {},
}

func (wc WordClass) Validate() error {
	if _, ok := validWordClasses[wc]; !ok {
		return ErrInvalidWordClass
	}
	return nil
}

// Dialect represents the Banjar dialect variant.
type Dialect string

const (
	DialectHulu Dialect = "hulu"
)

func (d Dialect) Validate() error {
	if d != DialectHulu {
		return ErrInvalidDialect
	}
	return nil
}

// Source indicates the origin of a dictionary entry.
type Source string

const (
	SourceSeeded      Source = "seeded"
	SourceContributed Source = "contributed"
	SourceAIGenerated Source = "ai_generated"
)

func (s Source) Validate() error {
	switch s {
	case SourceSeeded, SourceContributed, SourceAIGenerated:
		return nil
	}
	return ErrInvalidSource
}

// WordStatus represents the lifecycle state of a word.
type WordStatus string

const (
	WordStatusActive     WordStatus = "active"
	WordStatusDeprecated WordStatus = "deprecated"
)

func (ws WordStatus) Validate() error {
	switch ws {
	case WordStatusActive, WordStatusDeprecated:
		return nil
	}
	return ErrInvalidStatus
}
