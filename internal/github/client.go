package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Akash18272001/codepilot/internal/config"
	"github.com/Akash18272001/codepilot/internal/logger"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) CreatePullRequest(ctx context.Context, issueTitle, patchContent string) (string, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", c.cfg.GitHubOwner, c.cfg.GitHubRepo)

	// 1. Get Main Branch SHA
	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+"/git/ref/heads/main", nil)
	req.Header.Set("Authorization", "Bearer "+c.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error fetching ref: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github error (%d): %s", resp.StatusCode, string(body))
	}

	var refData struct {
		Object struct {
			Sha string `json:"sha"`
		} `json:"object"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&refData)

	// 2. Create Branch
	branchName := fmt.Sprintf("ai-remediation-%d", time.Now().Unix())
	refBody, _ := json.Marshal(map[string]string{
		"ref": "refs/heads/" + branchName,
		"sha": refData.Object.Sha,
	})

	req, _ = http.NewRequestWithContext(ctx, "POST", baseURL+"/git/refs", bytes.NewBuffer(refBody))
	req.Header.Set("Authorization", "Bearer "+c.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create branch (%d): %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	// 3. Commit Patch File
	encodedContent := base64.StdEncoding.EncodeToString([]byte(patchContent))
	fileBody, _ := json.Marshal(map[string]string{
		"message": "AI Agent: automated remediation for " + issueTitle,
		"content": encodedContent,
		"branch":  branchName,
	})

	req, _ = http.NewRequestWithContext(ctx, "PUT", baseURL+"/contents/patches/remediation.diff", bytes.NewBuffer(fileBody))
	req.Header.Set("Authorization", "Bearer "+c.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil || (resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK) {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to commit file (%d): %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	// 4. Open PR
	prBody, _ := json.Marshal(map[string]string{
		"title": "[AI Agent] Automated Fix: " + issueTitle,
		"head":  branchName,
		"base":  "main",
		"body":  "Autonomous AI Software Engineering Agent generated this fix following observability alert telemetry.",
	})

	req, _ = http.NewRequestWithContext(ctx, "POST", baseURL+"/pulls", bytes.NewBuffer(prBody))
	req.Header.Set("Authorization", "Bearer "+c.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to open PR (%d): %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	var prRes struct {
		HTMLURL string `json:"html_url"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&prRes)

	logger.Log.Info().Str("pr_url", prRes.HTMLURL).Msg("Successfully opened GitHub Pull Request")
	return prRes.HTMLURL, nil
}