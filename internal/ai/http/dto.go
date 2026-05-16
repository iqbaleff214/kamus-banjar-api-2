package aihttp

import "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"

type translateRequest struct {
	Text    string `json:"text"`
	Context string `json:"context"`
}

type translationResponse struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
	Dialect     string `json:"dialect"`
	Model       string `json:"model"`
	Confidence  string `json:"confidence"`
	Notes       string `json:"notes,omitempty"`
}

func toTranslationResponse(r *commands.TranslationResult) translationResponse {
	return translationResponse{
		Original:    r.Original,
		Translation: r.Translation,
		Dialect:     r.Dialect,
		Model:       r.Model,
		Confidence:  r.Confidence,
		Notes:       r.Notes,
	}
}
