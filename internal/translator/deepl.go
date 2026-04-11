package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	translate "github.com/OwO-Network/DeepLX/translate"
)

type DeepLTranslator struct {
	url string
}

func NewDeepLTranslator(url string) *DeepLTranslator {
	return &DeepLTranslator{
		url: url,
	}
}

func (d *DeepLTranslator) Name() string {
	return "DeepL"
}

func (d *DeepLTranslator) Translate(input, sourceLang, targetLang string) (*TranslationResult, error) {
	if d.url != "" {
		return d.translateCustom(input, sourceLang, targetLang)
	}

	result, err := translate.TranslateByDeepLX(sourceLang, targetLang, input, "html", "", "")
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Text:         result.Data,
		Alternatives: result.Alternatives,
	}, nil
}

func (d *DeepLTranslator) translateCustom(input, sourceLang, targetLang string) (*TranslationResult, error) {
	apiUrl := d.url
	if !strings.HasSuffix(apiUrl, "/") {
		apiUrl += "/translate"
	} else {
		apiUrl += "translate"
	}

	payload := map[string]string{
		"text":        input,
		"source_lang": sourceLang,
		"target_lang": targetLang,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DeepLX custom instance returned status %d", resp.StatusCode)
	}

	var data struct {
		Code         int      `json:"code"`
		Data         string   `json:"data"`
		Alternatives []string `json:"alternatives"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &TranslationResult{
		Text:         data.Data,
		Alternatives: data.Alternatives,
	}, nil
}
