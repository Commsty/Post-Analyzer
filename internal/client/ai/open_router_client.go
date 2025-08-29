package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultModel string = "deepseek/deepseek-chat-v3.1:free"
	defaultURL   string = "https://openrouter.ai/api/v1/chat/completions"
)

type OpenRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type openRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Reasoning   Reasoning `json:"reasoning"`
	Verbosity   string    `json:"verbosity"`
	Temperature float32   `json:"temperature"`
	ToPP        float32   `json:"top_p"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Reasoning struct {
	Enabled bool `json:"enabled"`
}

type openRouterResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}

func NewOpenRouterClient(apikey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:  apikey,
		baseURL: defaultURL,
		httpClient: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}
}

func (c *OpenRouterClient) AnalyzeText(ctx context.Context, text, sysPrompt string) (string, error) {

	if err := ctx.Err(); err != nil {
		return "", err
	}

	userPrompt := "Проанализируй следующие посты из Telegram-каналов и выдай краткую выжимку по заданным правилам:\n" + text

	reqBody := openRouterRequest{
		Model: defaultModel,
		Messages: []message{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt},
		},
		Reasoning: Reasoning{
			Enabled: false,
		},
		Verbosity:   "low",
		Temperature: 0.1,
		ToPP:        0.5,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON request marshalling error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("reqest creating error: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	if err := ctx.Err(); err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed: %d", resp.StatusCode)
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response failed: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response")
}
