package dictionaryhttp

import (
	"time"

	"github.com/google/uuid"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
)

type definitionResponse struct {
	ID        uuid.UUID `json:"id"`
	Meaning   string    `json:"meaning"`
	SortOrder int       `json:"sort_order"`
	Source    string    `json:"source"`
	Upvotes   int       `json:"upvotes"`
	Downvotes int       `json:"downvotes"`
	NetScore  int       `json:"net_score"`
}

type exampleResponse struct {
	ID                    uuid.UUID `json:"id"`
	BanjarSentence        string    `json:"banjar_sentence"`
	IndonesianTranslation string    `json:"indonesian_translation"`
	Source                string    `json:"source"`
}

type wordResponse struct {
	ID                uuid.UUID            `json:"id"`
	Banjar            string               `json:"banjar"`
	BanjarSyllabified string               `json:"banjar_syllabified,omitempty"`
	Dialect           string               `json:"dialect"`
	WordClass         string               `json:"word_class"`
	HomonymNumber     int                  `json:"homonym_number"`
	IsRoot            bool                 `json:"is_root"`
	RootWordID        *uuid.UUID           `json:"root_word_id,omitempty"`
	Definitions       []definitionResponse `json:"definitions"`
	Examples          []exampleResponse    `json:"examples"`
	Status            string               `json:"status"`
	Source            string               `json:"source"`
	SourceReference   string               `json:"source_reference,omitempty"`
	CreatedBy         *uuid.UUID           `json:"created_by,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

type wordSummaryResponse struct {
	ID            uuid.UUID `json:"id"`
	Banjar        string    `json:"banjar"`
	WordClass     string    `json:"word_class"`
	Dialect       string    `json:"dialect"`
	HomonymNumber int       `json:"homonym_number"`
	IsRoot        bool      `json:"is_root"`
	Status        string    `json:"status"`
	Source        string    `json:"source"`
	FirstMeaning  string    `json:"first_meaning,omitempty"`
}

type wordInputRequest struct {
	Banjar            string            `json:"banjar"`
	BanjarSyllabified string            `json:"banjar_syllabified"`
	WordClass         string            `json:"word_class"`
	Dialect           string            `json:"dialect"`
	HomonymNumber     int               `json:"homonym_number"`
	IsRoot            bool              `json:"is_root"`
	RootWordID        *uuid.UUID        `json:"root_word_id"`
	Definitions       []defInputReq     `json:"definitions"`
	Examples          []exampleInputReq `json:"examples"`
	SourceReference   string            `json:"source_reference"`
}

type defInputReq struct {
	Meaning   string `json:"meaning"`
	SortOrder int    `json:"sort_order"`
}

type exampleInputReq struct {
	BanjarSentence        string `json:"banjar_sentence"`
	IndonesianTranslation string `json:"indonesian_translation"`
}

func toWordResponse(w *domain.Word) wordResponse {
	defs := make([]definitionResponse, 0, len(w.Definitions))
	for _, d := range w.Definitions {
		defs = append(defs, definitionResponse{
			ID: d.ID, Meaning: d.Meaning, SortOrder: d.SortOrder,
			Source: string(d.Source), Upvotes: d.Upvotes, Downvotes: d.Downvotes, NetScore: d.NetScore(),
		})
	}
	exs := make([]exampleResponse, 0, len(w.Examples))
	for _, e := range w.Examples {
		exs = append(exs, exampleResponse{
			ID: e.ID, BanjarSentence: e.BanjarSentence,
			IndonesianTranslation: e.IndonesianTranslation, Source: string(e.Source),
		})
	}
	return wordResponse{
		ID: w.ID, Banjar: w.Banjar, BanjarSyllabified: w.BanjarSyllabified,
		Dialect: string(w.Dialect), WordClass: string(w.WordClass),
		HomonymNumber: w.HomonymNumber, IsRoot: w.IsRoot, RootWordID: w.RootWordID,
		Definitions: defs, Examples: exs, Status: string(w.Status), Source: string(w.Source),
		SourceReference: w.SourceReference, CreatedBy: w.CreatedBy,
		CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt,
	}
}

func toWordSummaryResponse(w *domain.Word) wordSummaryResponse {
	r := wordSummaryResponse{
		ID: w.ID, Banjar: w.Banjar, WordClass: string(w.WordClass),
		Dialect: string(w.Dialect), HomonymNumber: w.HomonymNumber,
		IsRoot: w.IsRoot, Status: string(w.Status), Source: string(w.Source),
	}
	if len(w.Definitions) > 0 {
		r.FirstMeaning = w.Definitions[0].Meaning
	}
	return r
}
