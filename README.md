# Task Management Service

Сервис для управления задачами с асинхронной обработкой.

## Возможности

- ✅ Создание задач через REST API
- ✅ Получение списка задач с пагинацией
- ✅ Получение задачи по ID
- ✅ Обновление статуса задачи
- ✅ Асинхронная обработка задач через очередь
- ✅ CORS поддержка
- ✅ Логирование
- ✅ Docker контейнеризация

## Архитектура

Проект следует принципам Clean Architecture:

- **Handler** - HTTP слой (`internal/delivery/handlers/`)
- **Service** - Бизнес-логика (`internal/domain/services/`)
- **Repository** - Работа с БД (`internal/domain/repositories/`)
- **Worker** - Асинхронная обработка (`internal/workers/`)

## API Endpoints

### POST /tasks
Создание новой задачи
```json
{
  "title": "Название задачи",
  "description": "Описание задачи"
}
```

### GET /tasks
Получение списка задач с пагинацией
```
GET /tasks?page=1&amount=10
```

### GET /tasks/{id}
Получение задачи по ID

### PATCH /tasks/{id}/status
Обновление статуса задачи
```
PATCH /tasks/{id}/status?status=processing
```

## Статусы задач

- `created` - Задача создана
- `processing` - Задача обрабатывается
- `done` - Задача выполнена
- `failed` - Задача завершилась с ошибкой

## Запуск

### Локальная разработка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте файл `.env`:
```env
APP_ENV=local
SERVER_PORT=3333
METRIC_SERVER_PORT=9090
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=taskdb
POSTGRES_PORT=5432
SSL_MODE=disable
CONFIG_PATH=./configs/app
```

3. Запустите PostgreSQL:
```bash
docker-compose up -d postgres
```

4. Выполните миграции:
```bash
make docker-migrate-up
```

5. Запустите приложение:
```bash
make run
```

### Docker

Запуск всего стека:
```bash
make docker-run-all
```

## Использование клиента

Пример использования сгенерированного клиента:

```go
package main

import (
    "betera-tz/client"
    "betera-tz/internal/dto"
    "context"
    "log"
)

func main() {
    c, err := client.NewClient("http://localhost:3333")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Создание задачи
    createReq := dto.PostTasksJSONRequestBody{
        Title:       stringPtr("Тестовая задача"),
        Description: stringPtr("Описание задачи"),
    }

    resp, err := c.PostTasks(ctx, createReq)
    if err != nil {
        log.Printf("Ошибка: %v", err)
        return
    }
    defer resp.Body.Close()

    log.Printf("Статус: %d", resp.StatusCode)
}
```

## Асинхронная обработка

При создании задачи:

1. Задача сохраняется в БД со статусом `created`
2. Задача отправляется в очередь
3. Воркер обрабатывает задачу:
   - Меняет статус на `processing`
   - Выполняет обработку (имитация работы)
   - Меняет статус на `done`

## Мониторинг

- Логи записываются в файл `/logs/betera-tz.log`
- Поддержка разных уровней логирования по окружениям
- Структурированные логи с контекстом

## Тестирование

```bash
# Запуск тестов
make test

# Запуск с verbose
go test ./... -v
```

## Docker команды

```bash
# Сборка
make docker-build

# Запуск инфраструктуры
make docker-run-infra

# Запуск приложения
make docker-run-app

# Просмотр логов
make docker-app-logs

# Остановка
make docker-down
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `APP_ENV` | Окружение (local/dev/prod) | local |
| `SERVER_PORT` | Порт сервера | 3333 |
| `POSTGRES_USER` | Пользователь БД | postgres |
| `POSTGRES_PASSWORD` | Пароль БД | password |
| `POSTGRES_DB` | Имя БД | taskdb |
| `POSTGRES_PORT` | Порт БД | 5432 |
| `SSL_MODE` | SSL режим | disable |
| `CONFIG_PATH` | Путь к конфигу | ./configs/app |

## Структура проекта

```
├── cmd/app/                 # Точка входа
├── internal/
│   ├── app/                 # Инициализация приложения
│   ├── config/             # Конфигурация
│   ├── delivery/           # HTTP слой
│   │   ├── handlers/       # HTTP обработчики
│   │   └── server/         # HTTP сервер
│   ├── domain/             # Доменная логика
│   │   ├── models/         # Модели данных
│   │   ├── repositories/   # Репозитории
│   │   └── services/       # Сервисы
│   ├── workers/            # Воркеры
│   └── migrations/         # Миграции БД
├── pkg/                    # Общие пакеты
│   ├── errs/              # Обработка ошибок
│   ├── logger/            # Логирование
│   ├── queue/             # Очередь
│   └── storage/          # Хранилище
├── client/                # Сгенерированный клиент
├── examples/              # Примеры использования
└── configs/               # Конфигурационные файлы
```
