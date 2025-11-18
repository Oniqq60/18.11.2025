## Link Checker

### 1. Запуск сервиса
 - go mod tidy
 - go run ./cmd/server
 - Остановка Ctr + C
Сервер слушает :8080. Батчи сохраняются в ./data

### 2. API
POST `/submit`
Пример JSON:
{
  "links": ["google.com", "malformedlink.gg"]
}

#### POST `/report`
Пример JSON:
{
  "links_num": [1, 2, 3]
}

- Если хотя бы один ID отсутствует, сервер отвечает 404.

### 3. Архитектура
- `internal/storage` — файловое хранилище
- `internal/service/checker` — проверка ссылок
- `internal/service/report` — генерация PDF через `gofpdf` (единственная библиотека).
- `internal/handlers` — HTTP-обработчики
- `internal/router` — маршрутизация



Ну вроде максимально просто попытался сделать      