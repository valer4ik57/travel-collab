# Travel-Collab

Travel-Collab — дипломный веб-сервис для совместного планирования путешествий.
Проект содержит backend на Go, frontend на Vue 3, PostgreSQL 16 + PostGIS, карту Leaflet/OpenStreetMap и WebSocket-синхронизацию.

## Что уже реализовано в этой версии MVP+

- регистрация и вход по email/паролю;
- JWT-авторизация;
- опциональный вход через GitHub OAuth2;
- создание поездок;
- invite-код для приглашения;
- вступление в поездку по invite-коду;
- список участников поездки;
- интерактивная карта Leaflet;
- добавление POI кликом по карте;
- редактирование и удаление точек;
- синхронизация точек через WebSocket;
- групповые расходы;
- расчёт балансов и упрощённых переводов "кто кому должен";
- чат внутри поездки через WebSocket;
- Docker Compose для PostgreSQL + PostGIS.

## Структура проекта

```text
travel-collab/
├── backend/                 # Go backend
│   ├── cmd/main.go
│   └── internal/
│       ├── api/             # HTTP handlers
│       ├── auth/            # JWT, password hashing
│       ├── config/          # env config
│       ├── db/              # PostgreSQL connection
│       ├── middleware/      # auth/cors middleware
│       ├── models/          # DTO/models
│       ├── repository/      # SQL queries
│       └── ws/              # WebSocket hub/client
├── frontend/                # Vue 3 + Vite + TypeScript frontend
├── sql/init.sql             # DB schema
├── docker-compose.yml       # PostgreSQL + PostGIS
├── .env.example
└── README.md
```

## Требования для запуска

- Windows 10/11;
- Docker Desktop;
- Go 1.23+;
- Node.js 20+;
- GoLand для backend;
- VS Code или GoLand для frontend.

## Быстрый запуск

### 1. Распаковать архив

Распакуй проект в удобную папку, например:

```powershell
C:\Projects\travel-collab
```

### 2. Запустить базу данных

В корне проекта:

```powershell
docker compose up -d
```

Проверить контейнер:

```powershell
docker ps
```

Если контейнер стартовал первый раз, `sql/init.sql` автоматически создаст таблицы и расширения `pgcrypto` и `postgis`.

### 3. Создать `.env`

Скопируй `.env.example` в `.env`:

```powershell
copy .env.example .env
```

Для локального запуска можно оставить значения по умолчанию.

### 4. Запустить backend

```powershell
cd backend
go mod tidy
go run ./cmd/main.go
```

Backend должен запуститься на:

```text
http://localhost:8080
```

Проверка:

```powershell
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

### 5. Запустить frontend

В новом терминале:

```powershell
cd frontend
npm install
npm run dev
```

Frontend откроется на:

```text
http://localhost:5173
```

## Как проверить основной сценарий

1. Открой `http://localhost:5173`.
2. Зарегистрируй пользователя.
3. Создай поездку.
4. Открой страницу поездки.
5. Кликни по карте и добавь точку.
6. Открой эту же поездку во второй вкладке или в другом браузере.
7. Добавь/измени точку — изменение должно прийти через WebSocket.
8. Добавь расход и проверь блок балансов.
9. Отправь сообщение в чат.

## Проверка invite-кода

1. Создай второго пользователя через другой браузер или режим инкогнито.
2. Скопируй invite-код из первой поездки.
3. На странице `/trips` второго пользователя введи invite-код.
4. Второй пользователь должен попасть в поездку.

## GitHub OAuth2

GitHub OAuth2 в проекте реализован как дополнительная функция. Для обычной демонстрации диплома он не нужен.

Чтобы включить:

1. Создай OAuth App в GitHub Developer Settings.
2. Callback URL укажи:

```text
http://localhost:8080/api/v1/auth/github/callback
```

3. Заполни в `.env`:

```env
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...
GITHUB_CALLBACK_URL=http://localhost:8080/api/v1/auth/github/callback
```

Если эти значения пустые, кнопка GitHub вернёт ошибку о том, что OAuth не настроен.

## Важные замечания

- Локально используется HTTP/WS. Для реального развёртывания HTTPS/WSS обычно включаются через reverse proxy, например Nginx или Caddy.
- Роль `viewer` заложена в БД, но основной MVP использует `owner` и `editor`.
- База запускается через Docker Compose, чтобы не устанавливать PostGIS вручную на Windows.
- Если меняешь `sql/init.sql`, а контейнер уже был создан, нужно пересоздать volume:

```powershell
docker compose down -v
docker compose up -d
```

## Что объяснять на защите

Краткая логика проекта:

- Backend реализует REST API для авторизации, поездок, точек, расходов и сообщений.
- JWT используется для проверки пользователя.
- PostgreSQL хранит основные данные, а PostGIS хранит географические координаты точек.
- Leaflet отображает карту и маркеры на frontend.
- WebSocket используется для событий реального времени: новые точки, изменение точек, чат, вступление участников и расходы.
- Расчёт балансов выполняется на backend: плательщик получает положительный баланс, участники разделения получают отрицательный баланс.

## Частые проблемы

### Backend не подключается к БД

Проверь, что Docker запущен и контейнер работает:

```powershell
docker ps
```

Проверь `DATABASE_URL` в `.env`.

### Frontend не видит backend

Убедись, что backend запущен на `localhost:8080`, а frontend на `localhost:5173`.
Vite proxy уже настроен в `frontend/vite.config.ts`.

### PostGIS extension error

Используй именно Docker image:

```text
postgis/postgis:16-3.4
```

В обычном PostgreSQL без PostGIS таблица `locations` не создастся.
