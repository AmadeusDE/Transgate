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

type OpenAITranslator struct {
	apiKey  string
	url     string
	model   string
	prompts config.Prompts
}

func NewOpenAITranslator(apiKey, url, model string, prompts config.Prompts) *OpenAITranslator {
	if url == "" {
		url = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	return &OpenAITranslator{
		apiKey:  apiKey,
		url:     url,
		model:   model,
		prompts: prompts,
	}
}

func (o *OpenAITranslator) Name() string {
	return o.model
}

func (o *OpenAITranslator) Request(ctx context.Context, input string) (string, error) {
	url := o.url + "/chat/completions"

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "user", "content": input},
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from openai")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func (o *OpenAITranslator) Translate(input, sourceLang, targetLang string) (*TranslationResult, error) {
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

func (o *OpenAITranslator) Detect(input string) (string, float64, error) {
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
