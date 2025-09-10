package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"post-analyzer/internal/adapters/openrouter"
	"post-analyzer/internal/adapters/telegram/bot"
	"post-analyzer/internal/adapters/telegram/user"
	"post-analyzer/internal/controllers"
	database "post-analyzer/internal/infrastructure/db"
	"post-analyzer/internal/infrastructure/notifier"
	"post-analyzer/internal/infrastructure/repository"
	"post-analyzer/internal/infrastructure/scheduler"
	"post-analyzer/internal/usecase"

	tgbot "github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("Failed to load startup configuration:\n" + err.Error())
	}
}

func main() {
	// loading configuration
	var token string = os.Getenv("TELEGRAM_BOT_TOKEN")
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatalf("Failed to get appID: %v", err)
	}
	var appHash string = os.Getenv("APP_HASH")
	var sessionPath string = os.Getenv("SESSION_PATH")
	var openRouterApiKey string = os.Getenv("OPEN_ROUTER_API_KEY")

	// db connection && check-up
	var db *pgxpool.Pool
	connectionCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if db, err = database.Init(connectionCtx); err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	if err = database.EnsureSchema(connectionCtx, db); err != nil {
		log.Fatalf("Failed to create databse tables: %v", err)
	}

	// repository
	repo := repository.NewSubscriptionRepository(db)

	// bot registartion
	botHandler, err := tgbot.New(token)
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
	aiClient := openrouter.NewOpenRouterClient(openRouterApiKey)

	// scheduler
	scheduler := scheduler.NewScheduler()

	// notifier
	notifier := notifier.NewNotifier(botClient)

	// usecase manager
	ucManager := usecase.NewUseCaseManager(userClient, repo, scheduler, aiClient, notifier)

	// bot messages handler
	handler := controllers.NewBotController(ucManager)

	// bot commands registration
	botHandler.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, handler.StartHandler)
	botHandler.RegisterHandler(tgbot.HandlerTypeMessageText, "/monitor", tgbot.MatchTypePrefix, handler.MonitorHandler)

	scheduler.Start()
	botHandler.Start(context.Background())
}
