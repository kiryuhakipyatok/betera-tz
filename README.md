# Task Management Service

Сервис для управления задачами с асинхронной обработкой.

## Возможности

- ✅ Создание задач через REST API
- ✅ Получение списка задач с пагинацией и фильтром по статусу
- ✅ Получение задачи по ID
- ✅ Обновление статуса задачи
- ✅ Асинхронная обработка задач через очередь
- ✅ CORS поддержка
- ✅ Логирование с использованием ELK стека
- ✅ Docker контейнеризация
- ✅ Swagger документация API
- ✅ Unit тесты для сервисного слоя
- ✅ Мониторинг при помощи Prometheus и Grafana

## Архитектура

Проект следует принципам Clean Architecture:

- **Handler** - HTTP слой (`internal/delivery/handlers/`)
- **Service** - Бизнес-логика (`internal/domain/services/`)
- **Repository** - Работа с БД (`internal/domain/repositories/`)
- **Worker** - Асинхронная обработка (`internal/workers/`)

## API Endpoints

### POST api/v1/tasks
Создание новой задачи
```json
{
  "title": "Название задачи",
  "description": "Описание задачи"
}
```

### GET api/v1/tasks
Получение списка задач с пагинацией и фильтром по статусу
```
GET api/v1/tasks?page=1&amount=10&statusFilter=done
```

### GET api/v1/tasks/{id}
Получение задачи по ID
```
GET api/v1/tasks/{id}
```

### PATCH api/v1/tasks/{id}/status
Обновление статуса задачи
```
PATCH api/v1/tasks/{id}/status?status=done
```

### GET /swagger
Swagger UI документация API
```
GET /swagger
```

## Статусы задач

- `created` - Задача создана
- `processing` - Задача обрабатывается
- `done` - Задача выполнена

## Запуск

### Локальная разработка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте файл `.env`:
```env
CONFIG_PATH = /configs/app
APP_ENV = dev
POSTGRES_PORT = 5432
SERVER_PORT = 3333
POSTGRES_PASSWORD = root
METRIC_SERVER_PORT = 3334
POSTGRES_HOST = postgres
SSL_MODE = disable
POSTGRES_USER = use
POSTGRES_DB = betera-tz
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${SSL_MODE}
GOOSE_MIGRATIONS_DIR = /migrations
MIGRATIONS_PATH=./internal/migrations
GF_SECURITY_ADMIN_PASSWORD = root
GRAFANA_PORT = 3000
PROMETHEUS_PORT = 9090
ELASTICSEARCH_PASSWORD = root
ELASTICSEARCH_USER = user
ELASTICSEARCH_PORT = 9200
KAFKA_PORT = 9092
KIBANA_PORT = 5601
```

3. В /logs создайте файл betera.log, а в /internal/domain/services test.log:

4. Запустите приложение с инфраструктурой:
```bash
make docker-run-all
```

5. Выполните миграции в другом терминале:
```bash
make docker-migrate-up
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
# Запуск всех тестов
make test

# Запуск unit тестов для сервисного слоя
make test-unit
```

## Docker команды

```bash
# Сборка
make docker-build

# Запуск инфраструктуры
make docker-run-infra

# Запуск elk стека
make docker-run-elk

# Запуск приложения
make docker-run-app

# Запуск мониторинга
make docker-run-monitoring

# Просмотр логов
make docker-app-logs

# Запуск всего сразу
make docker-run-all

# Остановка
make docker-down
```

## Переменные окружения

| Переменная                 | Описание                                | По умолчанию |
|-----------------------------|-----------------------------------------|--------------|
| `APP_ENV`                   | Окружение (local/dev/prod)              | dev |
| `SERVER_PORT`               | Порт сервера                            | 3333 |
| `POSTGRES_USER`             | Пользователь БД                         | user |
| `POSTGRES_PASSWORD`         | Пароль БД                               | root |
| `POSTGRES_DB`               | Имя БД                                  | betera-tz |
| `POSTGRES_PORT`             | Порт БД                                 | 5432 |
| `POSTGRES_HOST`             | Хост БД                                 | localhost |
| `SSL_MODE`                  | SSL режим                               | disable |
| `CONFIG_PATH`               | Путь к конфигу                          | ./configs/app |
| `METRIC_SERVER_PORT`        | Порт метрик (Prometheus endpoint)       | 3334 |
| `GOOSE_DRIVER`              | Драйвер для goose миграций              | postgres |
| `GOOSE_DBSTRING`            | Строка подключения для goose            | postgresql://postgres:password@localhost:5432/taskdb sslmode=disable |
| `GOOSE_MIGRATIONS_DIR`      | Директория миграций для goose           | ./migrations |
| `MIGRATIONS_PATH`           | Локальный путь к миграциям              | ./internal/migrations |
| `GF_SECURITY_ADMIN_PASSWORD`| Пароль администратора Grafana           | root |
| `GRAFANA_PORT`              | Порт Grafana                            | 3000 |
| `PROMETHEUS_PORT`           | Порт Prometheus                         | 9090 |
| `ELASTICSEARCH_USER`        | Пользователь Elasticsearch              | user |
| `ELASTICSEARCH_PASSWORD`    | Пароль Elasticsearch                    | root |
| `ELASTICSEARCH_PORT`        | Порт Elasticsearch                      | 9200 |
| `KAFKA_PORT`                | Порт Kafka                              | 9092 |
| `KIBANA_PORT`               | Порт Kibana                             | 5601 |


