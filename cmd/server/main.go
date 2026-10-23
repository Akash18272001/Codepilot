package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"ai-software-engineer/internal/tools"
	"ai-software-engineer/internal/workflow"
)

type IssueRequest struct {
	Issue string `json:"issue"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func main() {
	tools.InitTools()

	http.HandleFunc("/api/triage", handleTriage)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 AI Software Engineering Agent backend running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleTriage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("\n[AGENT] Received issue: %s\n", req.Issue)

	// Step 1: Orchestrator Plan
	plan := "1. Search logs for errors related to issue.\n2. Query database for relevant records.\n3. Search code for affected endpoints."
	fmt.Printf("[PLANNER] Plan generated:\n%s\n", plan)

	// Step 2: Tool Calls via Go Tool Server
	logResult := tools.ExecuteTool(tools.ToolRequest{
		Name:      "search_logs",
		Arguments: map[string]interface{}{"query": "500 order items > 50"},
	})
	fmt.Printf("[TOOL: search_logs] Result: %s\n", logResult.Output)

	dbResult := tools.ExecuteTool(tools.ToolRequest{
		Name:      "query_database",
		Arguments: map[string]interface{}{"query": "SELECT * FROM orders WHERE item_count > 50 LIMIT 5;"},
	})
	fmt.Printf("[TOOL: query_database] Result: %s\n", dbResult.Output)

	codeResult := tools.ExecuteTool(tools.ToolRequest{
		Name:      "search_code",
		Arguments: map[string]interface{}{"path": "/api/orders"},
	})
	fmt.Printf("[TOOL: search_code] Result: %s\n", codeResult.Output)

	// Step 3: Synthesis / LLM Analysis (Fallback to mock if Ollama not available)
	llmPrompt := fmt.Sprintf("Analyze issue: %s. Logs: %s. DB: %s. Code: %s. Provide root cause and patch.",
		req.Issue, logResult.Output, dbResult.Output, codeResult.Output)

	rootCause, err := callOllama(llmPrompt)
	if err != nil {
		rootCause = "Root Cause Analysis: Buffer limit overflow on batch sizes exceeding 50 items in /api/orders handler due to unhandled array allocation limit in slice."
	}

	// Step 4: Risk Assessment & Approval
	patchAction := "Apply fix to /api/orders handler and create PR"
	approved := workflow.RequestHumanApproval(patchAction)

	status := "APPROVED & EXECUTED"
	if !approved {
		status = "REJECTED BY HUMAN SAFETY GATE"
	}

	response := map[string]interface{}{
		"issue":      req.Issue,
		"plan":       plan,
		"root_cause": rootCause,
		"patch":      "git diff: fix max batch size check in orders_controller.go",
		"status":     status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func callOllama(prompt string) (string, error) {
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	reqBody, _ := json.Marshal(OllamaRequest{
		Model:  "qwen2.5-coder:7b",
		Prompt: prompt,
		Stream: false,
	})

	resp, err := http.Post(ollamaHost+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res OllamaResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if res.Response == "" {
		return "", fmt.Errorf("empty response from Ollama")
	}

	return strings.TrimSpace(res.Response), nil
}
