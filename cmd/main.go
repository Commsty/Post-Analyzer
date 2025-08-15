package main

import (
	"context"
	"log"
	"os"
	"post-analyzer/internal/client/ai"
	"post-analyzer/internal/client/telegram/bot"
	"post-analyzer/internal/client/telegram/user"
	"post-analyzer/internal/handlers"
	"post-analyzer/internal/repository"
	"post-analyzer/internal/scheduler"
	"post-analyzer/internal/service"
	"strconv"

	tgBot "github.com/go-telegram/bot"
	"github.com/joho/godotenv"
)

func main() {
	// loading env variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load env variables: %v", err)
	}

	var token string = os.Getenv("TELEGRAM_BOT_TOKEN")
	var appID, err = strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatalf("Failed to get appID: %v", err)
	}
	var appHash string = os.Getenv("APP_HASH")
	var storagePath string = os.Getenv("STORAGE_PATH")
	var sessionPath string = os.Getenv("SESSION_PATH")
	var openRouterApiKey string = os.Getenv("OPEN_ROUTER_API_KEY")

	// repository
	repo, err := repository.NewChannelStorage(storagePath)
	if err != nil {
		log.Fatalf("Failed to access data storage: %v", err)
	}

	// bot registartion
	botHandler, err := tgBot.New(token)
	if err != nil {
		log.Fatalf("Failed to register bot: %v", err)
	}

	// telegram clients
	botClient := bot.NewTelegramBotClient(botHandler)
	userClient, err := user.NewTelegramUserClient(appID, appHash, sessionPath)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}

	// OpenRouter client
	aiClient := ai.NewOpenRouterClient(openRouterApiKey)

	// scheduler
	scheduler := scheduler.NewScheduler()

	// internal services
	telegramProvider := service.NewTelegramProvider(userClient, botClient)
	analysisProvider := service.NewAnalysisProvider(aiClient)
	service := service.NewChannelService(repo, scheduler, analysisProvider, telegramProvider)

	// bot messages handler
	handler := handlers.NewBotHandler(service)

	// bot commands registration
	botHandler.RegisterHandler(tgBot.HandlerTypeMessageText, "/start", tgBot.MatchTypeExact, handler.StartHandler)
	botHandler.RegisterHandler(tgBot.HandlerTypeMessageText, "/monitor", tgBot.MatchTypePrefix, handler.MonitorHandler)

	scheduler.Start()
	botHandler.Start(context.Background())
}
