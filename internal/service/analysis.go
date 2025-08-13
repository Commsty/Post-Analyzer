package service

import (
	"context"
	"post-analyzer/internal/client/ai"
)

const basePrompt string = `Проанализируй эти новости и выдели из них только самые важные для жителя Москвы.
           Твоя цель - отфильтровать минимум новостей, которые несут максимум пользы для человека, а ненужные выкинуть.
           Никаких вступлений, просто тезисный сжатый анализ.
           Ответ выводи в формате:

           1) Новость номер 1
             - Причины, по которым новость важна
          	 - Последствия, которые новость несет

           2) Новость номер 2
		   	 - ...
			 - ...`

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

	return a.aiClient.AnalyzeText(ctx, text, basePrompt)

}
