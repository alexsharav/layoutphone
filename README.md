## Запуск сервера

Перейдите в папку сервера:

```bash
cd server
```

Установите зависимости:

```bash
go mod download
```

Проверьте файл `.env`:

```env
PORT=8080
FRONTEND_URL=http://localhost:5173
APP_ENV=development
ROUND_DEFAULT_TIME=300
TOPIC_DEFAULT_TIME=60
```

Запустите сервер:

```bash
go run main.go
```

Сервер будет доступен по адресу:

```text
http://localhost:8080
```

## Запуск клиента

Откройте второй терминал и перейдите в папку клиента:

```bash
cd client
```

Установите зависимости:

```bash
npm install
```

Проверьте файл `.env`:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
VITE_APP_NAME=Layout Phone
VITE_ENABLE_MOCKS=false
```

Запустите клиент:

```bash
npm run dev
```

Клиент будет доступен по адресу:

```text
http://localhost:5173
```

## Структура проекта

```text
client/   фронтенд на Vue 3, TypeScript, Vite, Pinia и Vue Router
server/   сервер на Go, Gin и Gorilla WebSocket
additional/   дополнительные материалы проекта
```
