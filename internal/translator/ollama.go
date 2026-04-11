package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"unix-supremacist.github.io/transgate/internal/config"
)

type OllamaTranslator struct {
	url     string
	model   string
	prompts config.Prompts
}

func NewOllamaTranslator(url, model string, prompts config.Prompts) *OllamaTranslator {
	if url == "" {
		url = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}
	return &OllamaTranslator{
		url:     url,
		model:   model,
		prompts: prompts,
	}
}

func (o *OllamaTranslator) Name() string {
	return o.model
}

func (o *OllamaTranslator) Request(ctx context.Context, input string) (string, error) {
	apiURL := o.url + "/api/generate"

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":  o.model,
		"prompt": input,
		"stream": false,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return strings.TrimSpace(result.Response), nil
}

func (o *OllamaTranslator) Translate(input, sourceLang, targetLang string) (*TranslationResult, error) {
	ctx := context.Background()
	prompt := o.prompts.Translate
	if prompt == "" {
		prompt = "can you give me a direct translation of the below sentence from %s to %s, with no explanation or any other fluff\n%s"
	}
	prompt = fmt.Sprintf(prompt, sourceLang, targetLang, input)

	translated, err := o.Request(ctx, prompt)

	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Text: translated,
	}, nil
}

func (o *OllamaTranslator) Detect(input string) (string, float64, error) {
	ctx := context.Background()
	prompt := o.prompts.Detect
	if prompt == "" {
		prompt = "detect the language of the following text and return ONLY a JSON object with keys 'lang' (ISO 639-1 code) and 'confidence' (float 0.0-1.0), please ensure the JSON is valid and has no markdown formatting:\n%s"
	}
	prompt = fmt.Sprintf(prompt, input)

	respStr, err := o.Request(ctx, prompt)

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
