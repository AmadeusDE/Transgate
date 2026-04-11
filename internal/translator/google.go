package translator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type GoogleTranslator struct {
	client *http.Client
}

func NewGoogleTranslator() *GoogleTranslator {
	return &GoogleTranslator{
		client: &http.Client{},
	}
}

func (g *GoogleTranslator) Name() string {
	return "Google Translate"
}

func (g *GoogleTranslator) Translate(input, sourceLang, targetLang string) (*TranslationResult, error) {
	var result []interface{}
	var translated []string

	urlStr := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s",
		sourceLang,
		targetLang,
		url.QueryEscape(input),
	)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("google translate failed with status %d", res.StatusCode)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		data := result[0]
		for _, slice := range data.([]interface{}) {
			for _, translatedText := range slice.([]interface{}) {
				translated = append(translated, fmt.Sprintf("%v", translatedText))
				break
			}
		}
		return &TranslationResult{
			Text: strings.Join(translated, ""),
		}, nil
	}

	return nil, fmt.Errorf("no translation found")
}
