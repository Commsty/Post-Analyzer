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
	defaultModel string = "openai/gpt-oss-20b:free"
	defaultURL   string = "https://openrouter.ai/api/v1/chat/completions"
)

type OpenRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type openRouterRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

func (c *OpenRouterClient) AnalyzeText(ctx context.Context, text, prompt string) (string, error) {

	fullPrompt := fmt.Sprintf(prompt, text)

	reqBody := openRouterRequest{
		Model: defaultModel,
		Messages: []message{
			{Role: "user", Content: fullPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON request marshalling error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Reqest creating error: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Sending request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed: %d", resp.StatusCode)
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("Decoding response failed: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("Empty response")
}
