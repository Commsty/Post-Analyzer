package main

import (
	"context"
	"log"
	"time"

	"post-analyzer/config"
	"post-analyzer/internal/adapters/openrouter"
	"post-analyzer/internal/adapters/telegram/bot"
	"post-analyzer/internal/adapters/telegram/user"
	"post-analyzer/internal/controllers"
	dtb "post-analyzer/internal/infrastructure/db"
	"post-analyzer/internal/infrastructure/notifier"
	"post-analyzer/internal/infrastructure/repository"
	"post-analyzer/internal/infrastructure/scheduler"
	"post-analyzer/internal/usecase"

	tgbot "github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

const cfgPath = "config/config.yml"

func main() {
	// loading configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		panic("failed to load config:\n" + err.Error())
	}

	// db connection && check-up
	var db *pgxpool.Pool
	connectionCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if db, err = dtb.Init(connectionCtx, cfg); err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	if err = dtb.EnsureSchema(connectionCtx, db); err != nil {
		log.Fatalf("Failed to create databse tables: %v", err)
	}

	// repository
	repo := repository.NewSubscriptionRepository(db)

	// bot registartion
	botHandler, err := tgbot.New(cfg.API.Telegram.BotToken)
	if err != nil {
		log.Fatalf("Failed to register bot: %v", err)
	}

	// telegram clients
	botClient := bot.NewTelegramBotClient(botHandler)
	userClient, err := user.NewTelegramUserClient(cfg.API.Telegram.AppID, cfg.API.Telegram.AppHash, cfg.API.Telegram.SessionPath)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}

	// OpenRouter client
	aiClient := openrouter.NewOpenRouterClient(cfg.API.OpenRouter.APIKey)

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
