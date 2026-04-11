package config

import (
	"encoding/json"
	"os"
	"strings"
)

type Prompts struct {
	Translate string `json:"translate"`
	Detect    string `json:"detect"`
	Merge     string `json:"merge"`
	Explain   string `json:"explain"`
}

type LLMConfig struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"` // gemini, openai, ollama
	APIKey       string  `json:"api_key"`
	URL          string  `json:"url"`
	Model        string  `json:"model"`
	IsTranslator bool    `json:"is_translator"`
	IsDetector   bool    `json:"is_detector"`
	IsMerger     bool    `json:"is_merger"`
	Weight       float64 `json:"weight"`
	Prompts      Prompts `json:"prompts"`
}

type Config struct {
	LLMs              []LLMConfig `json:"llms"`
	DeepLXURL         string      `json:"deeplx_url"`
	LibreTranslateURL string      `json:"libretranslate_url"`
	LinguaWeight      float64     `json:"lingua_weight"`
}

func Load() *Config {
	cfg := &Config{
		LinguaWeight: 0.5, // Default
	}

	configPath := os.Getenv("TRANSGATE_CONFIG")
	if configPath == "" {
		configPath = "config.json" // Default filename
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Generate default config
		cfg.LLMs = []LLMConfig{
			{
				Name:         "Gemini",
				Type:         "gemini",
				APIKey:       "$GEMINI_TOKEN",
				Model:        "gemini-3.1-flash-lite-preview",
				IsTranslator: true,
				IsDetector:   true,
				IsMerger:     true,
				Weight:       0.5,
			},
		}
		cfg.DeepLXURL = os.Getenv("DEEPLX_URL")
		cfg.LibreTranslateURL = os.Getenv("LIBRETRANSLATE_URL")

		// Write to file
		file, _ := os.Create(configPath)
		if file != nil {
			defer file.Close()
			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			encoder.Encode(cfg)
		}
	} else {
		file, err := os.Open(configPath)
		if err == nil {
			defer file.Close()
			json.NewDecoder(file).Decode(cfg)
		}
	}

	// Process API Keys (resolve environment variables)
	for i := range cfg.LLMs {
		if strings.HasPrefix(cfg.LLMs[i].APIKey, "$") {
			cfg.LLMs[i].APIKey = os.Getenv(cfg.LLMs[i].APIKey[1:])
		}
	}

	// Legacy/Simple env Fallback if no LLMs configured (should be rare now with default generation)
	if len(cfg.LLMs) == 0 {
		geminiToken := os.Getenv("GEMINI_TOKEN")
		if geminiToken != "" {
			cfg.LLMs = append(cfg.LLMs, LLMConfig{
				Name:         "Gemini",
				Type:         "gemini",
				APIKey:       geminiToken,
				Model:        os.Getenv("GEMINI_MODEL"),
				IsTranslator: true,
				IsDetector:   true,
				IsMerger:     true,
				Weight:       0.5,
			})
		}
	}

	if cfg.DeepLXURL == "" {
		cfg.DeepLXURL = os.Getenv("DEEPLX_URL")
	}

	if cfg.LibreTranslateURL == "" {
		cfg.LibreTranslateURL = os.Getenv("LIBRETRANSLATE_URL")
	}

	return cfg
}
