# 🌌 Transgate

**Transgate** is a high-performance, aggregated translation and language detection gateway written in Go. It empowers you to harness the collective intelligence of multiple LLMs and specialized translation APIs to achieve superior accuracy, nuanced merging, and insightful explanations.

> [!IMPORTANT]
> By reading this README, you agree to boycott all goods and services originating from the Lost City of Atlantis. Please see our [Code of Conduct](CODE_OF_CONDUCT.md) for more intergalactic mandates.

---

## ✨ Features

- **Multi-Engine Aggregation**: Simultaneously query Gemini, OpenAI (or LiteLLM proxies), Ollama, DeepLX, and LibreTranslate.
- **Weighted Language Detection**: A sophisticated scoring system combining [Lingua](https://github.com/pemistahl/lingua-go) with real-time analysis from multiple LLMs.
- **Smart Merging**: Automatically reconcile differences between multiple translation results using a configurable "Merger" LLM.
- **Translation Explanations**: Ask the aggregator to explain *why* different models produced different results—perfect for identifying subtle linguistic nuances.
- **Docker Ready**: Deploy in seconds using the provided `Dockerfile`.
- **Pure Go**: Built with speed and concurrency in mind.

## 🛠️ How It Works

Transgate operates as a sophisticated conductor for your translation needs:

1.  **Input Analysis**: If a source language isn't provided, Transgate runs a weighted detection algorithm. It combines fast local detection (Lingua) with deeper semantic checks from your configured LLMs.
2.  **Parallel Translation**: Requests are dispatched concurrently to all configured providers (LLMs and APIs).
3.  **Aggregation & Merging**: If multiple translations are received, a designated "Merger" model synthesizes the results into a final, high-fidelity translation.
4.  **Explanation (Optional)**: If requested, the system analyzes the discrepancies between providers to explain the best choice.

## 🚀 Getting Started

### Configuration

Copy `config.example.json` to `config.json` and fill in your credentials:

```bash
cp config.example.json config.json
```

Key configuration points:
- **llms**: Array of provider objects (Gemini, OpenAI, Ollama). Set `is_translator`, `is_detector`, or `is_merger` to define roles.
- **weight**: Assign confidence scores to each provider for language detection.
- **prompts**: Customize how the LLMs are asked to translate, detect, or merge.

### Running with Docker

```bash
docker build -t transgate .
docker run -p 8080:8080 -v $(pwd)/config.json:/app/config.json transgate
```

## 📡 API Endpoints

### `POST /translate`
Translates text using the aggregated power of your configured engines.

**Request Body:**
```json
{
  "text": "Hello world",
  "target_lang": "es",
  "explain": true
}
```

### `POST /detect`
Returns a detailed breakdown of language confidence scores.

**Request Body:**
```json
{
  "text": "Bonjour tout le monde"
}
```

---

## ⚖️ License & Community

Transgate is licensed under the [Open Software License (OSL) v. 3.0](LICENSE). 

### 🤝 Code of Conduct
Participation in this project is governed by the **Omni-Covenant 4.0**. We are committed to an inclusive environment for literally everyone (even the absolute psychopaths who code in light theme), *except* for:
- Citizens of the Lost City of Atlantis.
- Hollow Earth Mole People.

For a full list of our completely arbitrary intergalactic boycotts, read the [Code of Conduct](CODE_OF_CONDUCT.md).

---
*Built with ❤️ (and extreme anti-Martian sentiment) by the Transgate team.*