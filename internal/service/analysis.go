package service

import (
	"context"
	"post-analyzer/internal/client/ai"
)

const basePrompt string = `Ты — российский эксперт по анализу новостей из Telegram-каналов. Твоя задача — прочитать посты, проанализировать их содержание, отфильтровать кликбейт, слухи, эмоциональный шум, провокации и неподтверждённую информацию. Оставить только факты, подкреплённые достоверными данными.

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

Начни анализ:

`

type analysisProvider interface {
	AnalyzeTextWithBasicPrompt(context.Context, string) (string, error)
}

type analysisService struct {
	aiClient *ai.OpenRouterClient
}

func NewAnalysisProvider(aiClient *ai.OpenRouterClient) analysisProvider {
	return &analysisService{
		aiClient: aiClient,
	}
}

func (a *analysisService) AnalyzeTextWithBasicPrompt(ctx context.Context, text string) (string, error) {

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	return a.aiClient.AnalyzeText(ctx, text, basePrompt)
}
