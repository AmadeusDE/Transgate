package translator

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Aggregator struct {
	translators []Translator
	merger      LLM
	detector    *Detector
}

func NewAggregator(merger LLM, detector *Detector, translators ...Translator) *Aggregator {
	return &Aggregator{
		translators: translators,
		merger:      merger,
		detector:    detector,
	}
}

type AggregatedResult struct {
	Original    string                        `json:"original"`
	SourceLang  string                        `json:"source_lang"`
	TargetLang  string                        `json:"target_lang"`
	Merged      string                        `json:"merged"`
	Explanation string                        `json:"explanation,omitempty"`
	Providers   map[string]*TranslationResult `json:"providers"`
}

func (a *Aggregator) Translate(ctx context.Context, input, targetLang string, explain bool) (*AggregatedResult, error) {
	sourceLang := a.detector.Detect(input)
	return a.translate(ctx, input, sourceLang, targetLang, explain)
}

func (a *Aggregator) TranslateWithSource(ctx context.Context, input, sourceLang, targetLang string, explain bool) (*AggregatedResult, error) {
	return a.translate(ctx, input, sourceLang, targetLang, explain)
}

func (a *Aggregator) translate(ctx context.Context, input, sourceLang, targetLang string, explain bool) (*AggregatedResult, error) {
	results := make(map[string]*TranslationResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, t := range a.translators {
		wg.Add(1)
		go func(tr Translator) {
			defer wg.Done()
			res, err := tr.Translate(input, sourceLang, targetLang)
			if err == nil && res != nil && res.Text != "" {
				mu.Lock()
				results[tr.Name()] = res
				mu.Unlock()
			}
		}(t)
	}

	wg.Wait()

	if len(results) == 0 {
		return nil, fmt.Errorf("all translators failed")
	}

	// Prepare for merging
	var sumTranslations strings.Builder
	for name, res := range results {
		sumTranslations.WriteString(fmt.Sprintf("%s: %s\n", name, res.Text))
		for _, alt := range res.Alternatives {
			sumTranslations.WriteString(fmt.Sprintf("Alt: %s\n", alt))
		}
	}

	mergePrompt := fmt.Sprintf("can you concisely combine all these translations into the one that is most likely? if words are equally likely and is really important for context, like as in the words/set of words don't mean the same thing, you can use (word1/word2) with the parentheses format, but please omit words with the same meaning in context, like right/aren't we, please do not make any explanation or extra fluff, just give only the merged translation\n%s", sumTranslations.String())

	merged, err := a.merger.Request(ctx, mergePrompt)
	if err != nil {
		return nil, fmt.Errorf("merging failed: %w", err)
	}

	finalResult := &AggregatedResult{
		Original:   input,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Merged:     merged,
		Providers:  results,
	}

	if explain {
		explainPrompt := fmt.Sprintf("Original Sentence:\n%s\n\n%s\nThese are translations using a variety of translators, can you concisely summarize the meaning of the non-english words, explaining the translation differences and explain the meaning of the original sentence?, avoid generic explanations only provide piece by piece sentence explanations, Also try avoid naming the translators(deepl, gemini, etc), also try to keep explanations, within 400 characters", input, sumTranslations.String())
		explanation, err := a.merger.Request(ctx, explainPrompt)
		if err == nil {
			finalResult.Explanation = explanation
		}
	}

	return finalResult, nil
}

func (a *Aggregator) DetectScores(input string) map[string]float64 {
	return a.detector.DetectScores(input)
}
