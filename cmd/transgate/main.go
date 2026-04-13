package main

import (
	"log"
	"net/http"
	"os"

	"unix-supremacist.github.io/transgate/internal/api"
	"unix-supremacist.github.io/transgate/internal/config"
	"unix-supremacist.github.io/transgate/internal/translator"
)

func main() {
	cfg := config.Load()

	detector := translator.NewDetector(cfg.LinguaWeight)

	var translators []translator.Translator
	var merger translator.LLM

	for _, llmCfg := range cfg.LLMs {
		var llm translator.LLM
		var err error

		switch llmCfg.Type {
		case "gemini":
			llm, err = translator.NewGeminiTranslator(llmCfg.APIKey, llmCfg.Model, llmCfg.Prompts)
		case "openai":
			llm = translator.NewOpenAITranslator(llmCfg.APIKey, llmCfg.URL, llmCfg.Model, llmCfg.Prompts)

		case "ollama":
			llm = translator.NewOllamaTranslator(llmCfg.URL, llmCfg.Model, llmCfg.Prompts)
		default:

			log.Printf("Unknown LLM type: %s", llmCfg.Type)
			continue
		}

		if err != nil {
			log.Printf("Failed to initialize LLM %s: %v", llmCfg.Name, err)
			continue
		}

		if llmCfg.IsTranslator {
			translators = append(translators, llm)
		}
		if llmCfg.IsDetector {
			detector.AddLLMDetector(llm, llmCfg.Weight)
		}
		if llmCfg.IsMerger && merger == nil {
			merger = llm
		}
	}

	// Always add standard translators
	google := translator.NewGoogleTranslator()
	deepl := translator.NewDeepLTranslator(cfg.DeepLXURL)
	translators = append(translators, google, deepl)

	if cfg.LibreTranslateURL != "" {
		libre := translator.NewLibreTranslateTranslator(cfg.LibreTranslateURL)
		translators = append(translators, libre)
	}

	if merger == nil {
		log.Fatal("No merger LLM configured")
	}

	aggregator := translator.NewAggregator(merger, detector, translators...)

	server := api.NewServer(aggregator)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Transgate API on port %s", port)
	if err := http.ListenAndServe(":"+port, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
