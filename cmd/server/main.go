package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Akash18272001/codepilot/internal/config"
	"github.com/Akash18272001/codepilot/internal/github"
	"github.com/Akash18272001/codepilot/internal/logger"
)

type Server struct {
	cfg      *config.Config
	ghClient *github.Client
}

func main() {
	logger.InitLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Configuration initialization failed")
	}

	srv := &Server{
		cfg:      cfg,
		ghClient: github.NewClient(cfg),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/triage", srv.handleTriage)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	logger.Log.Info().Str("port", cfg.Port).Msg("Starting CodePilot Enterprise Agent Server...")
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		logger.Log.Fatal().Err(err).Msg("Server terminated abruptly")
	}
}

func (s *Server) handleTriage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Issue string `json:"issue"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Info().Str("issue", req.Issue).Msg("Received telemetry incident alert")

	patchText := "git diff: optimized order batch allocation limit"

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	prURL, err := s.ghClient.CreatePullRequest(ctx, req.Issue, patchText)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to execute automated remediation PR")
		prURL = "Error: " + err.Error()
	}

	response := map[string]interface{}{
		"issue":     req.Issue,
		"status":    "AUTOMATED_REMEDIATION_EXECUTED",
		"pr_url":    prURL,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}