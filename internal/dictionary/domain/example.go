package domain

import "github.com/google/uuid"

type Example struct {
	ID                  uuid.UUID
	BanjarSentence      string
	IndonesianTranslation string
	Source              Source
}

func NewExample(banjar, indonesian string, source Source) *Example {
	return &Example{
		ID:                    uuid.New(),
		BanjarSentence:        banjar,
		IndonesianTranslation: indonesian,
		Source:                source,
	}
}
