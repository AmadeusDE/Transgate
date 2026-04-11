package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"unix-supremacist.github.io/transgate/internal/config"
)

type GeminiTranslator struct {
	client  *genai.Client
	model   string
	prompts config.Prompts
}

func NewGeminiTranslator(apiKey, model string, prompts config.Prompts) (*GeminiTranslator, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiTranslator{
		client:  client,
		model:   model,
		prompts: prompts,
	}, nil
}

func (g *GeminiTranslator) Name() string {
	return g.model
}

func (g *GeminiTranslator) Translate(input, sourceLang, targetLang string) (*TranslationResult, error) {
	ctx := context.Background()
	model := g.client.GenerativeModel(g.model)

	prompt := g.prompts.Translate
	if prompt == "" {
		prompt = "can you give me a direct translation of the below sentence from %s to %s, with no explanation or any other fluff\n%s"
	}
	prompt = fmt.Sprintf(prompt, sourceLang, targetLang, input)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))

	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no response from gemini")
	}

	var builder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		builder.WriteString(fmt.Sprint(part))
	}

	return &TranslationResult{
		Text: strings.TrimSpace(builder.String()),
	}, nil
}

func (g *GeminiTranslator) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

// Request is a generic method for non-translation tasks (like explanation/merging)
func (g *GeminiTranslator) Request(ctx context.Context, input string) (string, error) {
	model := g.client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, genai.Text(input))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("no response from gemini")
	}

	var builder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		builder.WriteString(fmt.Sprint(part))
	}

	return strings.TrimSpace(builder.String()), nil
}

func (g *GeminiTranslator) Detect(input string) (string, float64, error) {
	ctx := context.Background()
	prompt := g.prompts.Detect
	if prompt == "" {
		prompt = "detect the language of the following text and return ONLY a JSON object with keys 'lang' (ISO 639-1 code) and 'confidence' (float 0.0-1.0), please ensure the JSON is valid and has no markdown formatting:\n%s"
	}
	prompt = fmt.Sprintf(prompt, input)

	respStr, err := g.Request(ctx, prompt)

	if err != nil {
		return "", 0, err
	}

	// Basic cleanup if LLM still includes markdown
	respStr = strings.TrimPrefix(respStr, "```json")
	respStr = strings.TrimSuffix(respStr, "```")
	respStr = strings.TrimSpace(respStr)

	var result struct {
		Lang       string  `json:"lang"`
		Confidence float64 `json:"confidence"`
	}

	err = json.Unmarshal([]byte(respStr), &result)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse detection result: %w", err)
	}

	return strings.ToLower(result.Lang), result.Confidence, nil
}
