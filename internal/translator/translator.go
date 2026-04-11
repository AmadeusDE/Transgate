package translator

import "context"

type TranslationResult struct {
	Text         string   `json:"text"`
	Alternatives []string `json:"alternatives,omitempty"`
}

type Translator interface {
	Translate(input, sourceLang, targetLang string) (*TranslationResult, error)
	Name() string
}

type LLM interface {
	Translator
	Request(ctx context.Context, input string) (string, error)
	Detect(input string) (string, float64, error)
}
