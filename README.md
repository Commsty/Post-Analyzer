# Post Analyzer

## Актуальные даты для проекта
Июль - Август 2025

## Описание проекта

Post Analyzer — это Telegram бот для мониторинга каналов и анализа новых постов с помощью AI. Бот отслеживает обновления в указанных каналах и автоматически анализирует содержание новых постов, отправляя краткие сводки пользователям.

Проект написан на Go с использованием чистой архитектуры (Clean Architecture) и показывает, как строить масштабируемые микросервисы.

## Что умеет бот

- **Мониторинг каналов** — отслеживает новые посты в Telegram каналах
- **AI-анализ** — анализирует содержание постов с помощью OpenRouter API
- **Планировщик** — настраиваемая проверка каналов по расписанию
- **Уведомления** — отправляет анализ постов подписчикам
- **База данных** — хранит подписки и состояние мониторинга

## Структура проекта

```
Post-Analyzer/
├── cmd/                     # Точка входа в приложение
│   └── main.go             # Основной файл приложения
├── config/                 # Конфигурация
│   ├── config.go          # Загрузка конфигурации
│   └── config.yml         # YAML конфигурация
├── internal/              # Внутренний код приложения
│   ├── adapters/          # Адаптеры внешних сервисов
│   │   ├── openrouter/   # Клиент OpenRouter API
│   │   └── telegram/     # Telegram клиенты (bot + user)
│   ├── controllers/       # Обработчики HTTP/bot запросов
│   ├── domain/           # Бизнес-логика
│   │   ├── entity/       # Сущности
│   │   ├── dto/          # Объекты передачи данных
│   │   ├── presenter/    # Презентеры
│   │   └── validation/   # Валидация
│   ├── infrastructure/   # Инфраструктурный слой
│   │   ├── db/          # Работа с базой данных
│   │   ├── notifier/    # Уведомления
│   │   ├── repository/  # Репозитории
│   │   └── scheduler/   # Планировщик задач
│   └── usecase/         # Слой бизнес-логики
├── .env.example          # Пример переменных окружения
├── docker-compose.yml    # Docker конфигурация
└── go.mod               # Go модули
```

## Технологический стек

- **Язык**: Go 1.25.1
- **База данных**: PostgreSQL с pgx/v5
- **Telegram API**: go-telegram/bot + gotd/td
- **AI**: OpenRouter API
- **Планировщик**: robfig/cron
- **Архитектура**: Clean Architecture

## Как запустить проект

### 1. Клонирование и установка зависимостей

```bash
git clone git@github.com:Commsty/Post-Analyzer.git
cd Post-Analyzer
go mod download
```

### 2. Настройка окружения

Скопируйте `.env.example` в `.env` и заполните данными:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
TG_BOT_TOKEN=your_telegram_bot_token
TG_APP_ID=your_telegram_app_id
TG_APP_HASH=your_telegram_app_hash
AUTH_PHONE=your_phone_number
AUTH_PASSWORD=your_password
SESSION_PATH=./session.json
OPENROUTER_API_KEY=your_openrouter_api_key
DB_PASSWORD=your_database_password
```

### 3. Настройка базы данных

Настройте подключение в config/config.yml

### 4. Запуск

```bash
# Разработка
go run cmd/main.go

# Сборка
go build -o bin/post-analyzer cmd/main.go
./bin/post-analyzer
```

### 5. С Docker

```bash
docker-compose up -d # для запуска базы данных
```

### 6. Тестирование

```bash
# Запуск всех тестов
go test ./... -v

# Запуск тестов с покрытием кода
go test -cover ./... -v

# Запуск тестов для конкретного пакета
go test ./cmd -v
go test ./config -v
go test ./internal/domain/dto -v
go test ./internal/domain/entity -v
go test ./internal/infrastructure/repository -v
```

### Основные команды

- `/start` — запуск бота и приветствие
- `/monitor @channel_name` — начать мониторинг канала
- `/monitor @channel_name 09:00` — мониторинг с уведомлениями в 09:00
