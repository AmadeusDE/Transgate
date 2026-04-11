package translator

import (
	"strings"

	"github.com/pemistahl/lingua-go"
)

type LLMDetector struct {
	LLM    LLM
	Weight float64
}

type Detector struct {
	linguaDetector lingua.LanguageDetector
	linguaWeight   float64
	llmDetectors   []LLMDetector
}

func NewDetector(linguaWeight float64) *Detector {
	languages := []lingua.Language{
		lingua.Arabic, lingua.Bulgarian, lingua.Chinese, lingua.Czech,
		lingua.Danish, lingua.Dutch, lingua.English, lingua.Estonian,
		lingua.Finnish, lingua.French, lingua.German, lingua.Greek,
		lingua.Hungarian, lingua.Indonesian, lingua.Italian, lingua.Japanese,
		lingua.Korean, lingua.Latin, lingua.Latvian, lingua.Lithuanian,
		lingua.Polish, lingua.Portuguese, lingua.Romanian, lingua.Russian,
		lingua.Slovak, lingua.Spanish, lingua.Swedish, lingua.Turkish,
		lingua.Ukrainian, lingua.Vietnamese,
	}

	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	return &Detector{
		linguaDetector: detector,
		linguaWeight:   linguaWeight,
		llmDetectors:   []LLMDetector{},
	}
}

func (d *Detector) AddLLMDetector(llm LLM, weight float64) {
	d.llmDetectors = append(d.llmDetectors, LLMDetector{
		LLM:    llm,
		Weight: weight,
	})
}

func (d *Detector) Detect(input string) string {
	scores := d.DetectScores(input)

	// Find best score
	bestLang := "unknown"
	maxScore := -1.0
	for lang, score := range scores {
		if score > maxScore {
			maxScore = score
			bestLang = lang
		}
	}

	return bestLang
}

func (d *Detector) DetectScores(input string) map[string]float64 {
	scores := make(map[string]float64)

	// Lingua detection
	confidence := d.linguaDetector.ComputeLanguageConfidenceValues(input)
	for _, v := range confidence {
		if v.Value() > 0 {
			langCode := strings.ToLower(v.Language().IsoCode639_1().String())
			scores[langCode] += v.Value() * d.linguaWeight
		}
	}

	// LLM detection
	for _, ld := range d.llmDetectors {
		lang, conf, err := ld.LLM.Detect(input)
		if err == nil {
			scores[lang] += conf * ld.Weight
		}
	}

	return scores
}
