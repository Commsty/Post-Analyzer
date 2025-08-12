package main

import (
	"context"
	"log"
	"os"
	"post-analyzer/internal/client/telegram/bot"
	"post-analyzer/internal/client/telegram/user"
	"post-analyzer/internal/handlers"
	"post-analyzer/internal/repository"
	"post-analyzer/internal/service"
	"strconv"

	tgBot "github.com/go-telegram/bot"
	"github.com/joho/godotenv"
)

func main() {

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

	botHandler, err := tgBot.New(token)
	if err != nil {
		log.Fatalf("Failed to register bot: %v", err)
	}
	userClient, err := user.NewTelegramUserClient(appID, appHash, sessionPath)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}
	botClient := bot.NewTelegramBotClient(botHandler)

	repo, err := repository.NewChannelStorage(storagePath)
	if err != nil {
		log.Fatalf("Failed to access data storage: %v", err)
	}
	serv := service.NewMonitoringService(repo, botClient, userClient)
	hand := handlers.NewBotHandler(serv)

	botHandler.RegisterHandler(tgBot.HandlerTypeMessageText, "/start", tgBot.MatchTypeExact, hand.StartHandler)
	botHandler.RegisterHandler(tgBot.HandlerTypeMessageText, "", tgBot.MatchTypePrefix, hand.LinkHandler)

	botHandler.Start(context.Background())
}
