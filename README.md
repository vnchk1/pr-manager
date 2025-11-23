# PR Manager Service - Инструкция по установке и запуску

Сервис для автоматического назначения ревьюеров на Pull Request'ы с управлением командами и пользователями.

## Предварительные требования

- Docker 20.10+ и Docker Compose
- Go 1.24
- Git

## Запуск

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/vnchk1/pr-manager.git <Ваша директория>
cd <Ваша директория>
```

2. **Запустите сервис:**
```bash
docker-compose up --build
```

Сервис будет доступен по адресу: `http://localhost:8080`

## Доступные команды Makefile

```bash
make build          # Сборка приложения
make run            # Запуск приложения
make test           # Запуск тестов
make lint           # Проверка кодстайла
make migrate-up     # Применить миграции
make migrate-down   # Откатить миграции
make docker-up      # Запуск через Docker
make docker-down    # Остановка Docker
make clean          # Очистка проекта
make fmt            # Форматирование кода
make health-check   # Проверка здоровья сервиса
```

## Проверка работоспособности

После запуска проверьте здоровье сервиса:

```bash
curl "http://localhost:8080/health"
```

Ожидаемый ответ: `{"status":"healthy","message":"Service is running"}`

## API Endpoints

### Команды
- `POST /team/add` - Создать команду
- `GET /team/get?team_name=name` - Получить команду

### Пользователи
- `POST /users/setIsActive` - Установить активность пользователя
- `GET /users/getReview?user_id=id` - Получить PR пользователя

### Pull Requests
- `POST /pullRequest/create` - Создать PR
- `POST /pullRequest/merge` - Смержить PR
- `POST /pullRequest/reassign` - Переназначить ревьювера

### Статистика
- `GET /stats/assignments` - Статистика назначений
- `GET /stats/user` - Статистика по пользователям

## Конфигурация

### Переменные окружения (указаны в docker-compose.yml)

```bash
DB_HOST=postgres          # Хост БД
DB_PORT=5432             # Порт БД  
DB_USER=postgres         # Пользователь БД
DB_PASSWORD=postgres     # Пароль БД
DB_NAME=pr_manager       # Имя БД
APP_PORT=8080            # Порт приложения
```

## Остановка сервиса

```bash
# Остановка с удалением томов
make clean

# Или только остановка
docker-compose down
```