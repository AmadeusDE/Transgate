package api

import (
	"encoding/json"
	"net/http"

	"unix-supremacist.github.io/transgate/internal/translator"
)

type Server struct {
	aggregator *translator.Aggregator
}

func NewServer(aggregator *translator.Aggregator) *Server {
	return &Server{aggregator: aggregator}
}

type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Explain    bool   `json:"explain"`
}

func (s *Server) HandleTranslate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.TargetLang == "" {
		req.TargetLang = "en"
	}

	var result *translator.AggregatedResult
	var err error
	if req.SourceLang != "" {
		result, err = s.aggregator.TranslateWithSource(r.Context(), req.Text, req.SourceLang, req.TargetLang, req.Explain)
	} else {
		result, err = s.aggregator.Translate(r.Context(), req.Text, req.TargetLang, req.Explain)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) HandleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	scores := s.aggregator.DetectScores(req.Text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/translate", s.HandleTranslate)
	mux.HandleFunc("/detect", s.HandleDetect)
	return mux
}
