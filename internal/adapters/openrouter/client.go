package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultModel string = "deepseek/deepseek-chat-v3.1:free"
	defaultURL   string = "https://openrouter.ai/api/v1/chat/completions"

	userPrompt string = "Проанализируй следующие посты из Telegram-каналов и выдай краткую выжимку по заданным правилам:\n"
	sysPrompt  string = `Ты — российский эксперт по анализу новостей из Telegram-каналов. Твоя задача — прочитать посты, проанализировать их содержание, отфильтровать кликбейт, слухи, эмоциональный шум, провокации и неподтверждённую информацию. Оставить только факты, подкреплённые достоверными данными.

		Сделай краткую выжимку из новостей по следующим правилам:
		- Только объективные, проверяемые факты.
		- Убери оценки, мнения, гипотезы, предположения.
		- Не используй метафоры, эмоциональные выражения, восклицания.
		- Сократи текст до 1 предложения на каждую новость.
		- Общее количество оставшихся новостей - не более 30% от изначального количества.
		- Если новость непроверенная, малозначимая или похожа на слух — проигнорируй её.
		- Группируй схожие новости в один пункт.
		- Формат: маркированный список, каждый пункт — одна важная новость Между новостями - пустая строка.
		- Анализ исключительно на русском языке.
		- Учти, твоя работа оценивается очень строго.

		Пример вывода:
		- [Краткая суть новости, только факты]

		- [Ещё одна новость, без лишних деталей]

		- [...]
		`
)

var (
	ErrJSONMarshalling = errors.New("marshalling failed")
	ErrJSONDecode      = errors.New("decoding response failed")

	ErrRequestCreation   = errors.New("req creation failed")
	ErrRequestDo         = errors.New("request sending failed")
	ErrRequestProcessing = errors.New("request processing failed")

	ErrEmptyResponse = errors.New("empty response")

	ErrTimeLimit = errors.New("openrouter req time limit reached")
)

type AnalysisService interface {
	AnalyzePosts(ctx context.Context, text string) (string, error)
}

type openRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewOpenRouterClient(apikey string) *openRouterClient {
	return &openRouterClient{
		apiKey:  apikey,
		baseURL: defaultURL,
		httpClient: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}
}

func (c openRouterClient) AnalyzePosts(ctx context.Context, text string) (string, error) {

	if err := ctx.Err(); err != nil {
		return "", ErrTimeLimit
	}

	userPrompt := userPrompt + text

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
		return "", fmt.Errorf("%w: %s", ErrJSONMarshalling, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrRequestCreation, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	if err := ctx.Err(); err != nil {
		return "", ErrTimeLimit
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrRequestDo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status code: %d", ErrRequestProcessing, resp.StatusCode)
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: %s", ErrJSONDecode, err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", ErrEmptyResponse
}
