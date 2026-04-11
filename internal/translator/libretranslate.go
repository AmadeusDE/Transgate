package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type LibreTranslateTranslator struct {
	url    string
	client *http.Client
}

func NewLibreTranslateTranslator(url string) *LibreTranslateTranslator {
	return &LibreTranslateTranslator{
		url:    url,
		client: &http.Client{},
	}
}

func (l *LibreTranslateTranslator) Name() string {
	return "LibreTranslate"
}

func (l *LibreTranslateTranslator) Translate(input, sourceLang, targetLang string) (*TranslationResult, error) {
	apiUrl := l.url
	if !strings.HasSuffix(apiUrl, "/") {
		apiUrl += "/translate"
	} else {
		apiUrl += "translate"
	}

	payload := map[string]string{
		"q":      input,
		"source": strings.ToLower(sourceLang),
		"target": strings.ToLower(targetLang),
		"format": "text",
	}
	body, _ := json.Marshal(payload)

	resp, err := l.client.Post(apiUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("LibreTranslate failed: %s (status %d)", errData.Error, resp.StatusCode)
	}

	var data struct {
		TranslatedText string `json:"translatedText"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &TranslationResult{
		Text: data.TranslatedText,
	}, nil
}
