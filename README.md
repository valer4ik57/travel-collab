# Travel-Collab

Travel-Collab — дипломный сервис для совместного планирования путешествий. Проект позволяет группе участников создать поездку, пригласить друзей по invite-коду, собрать маршруты по дням, отметить точки на карте, вести групповые расходы и общаться в чате в реальном времени.

Проект реализован как full-stack система:

- backend: Go, chi, pgx, JWT, WebSocket;
- база данных: PostgreSQL 16 + PostGIS;
- frontend: Vue 3 + Vite + TypeScript;
- карта: Leaflet + OpenStreetMap;
- мобильная версия: адаптивный web-интерфейс и Android APK через Capacitor;
- запуск базы: Docker Compose.

## Основные возможности

- регистрация и вход по email/паролю;
- JWT-авторизация и защита API;
- создание поездок;
- вступление в поездку по invite-коду;
- роли участников: `owner`, `editor`, `viewer`;
- интерактивная карта с точками интереса;
- маршруты поездки по дням;
- ручной порядок точек маршрута через drag-and-drop;
- построение маршрута во внешнем приложении 2ГИС;
- групповые расходы с несколькими плательщиками;
- распределение расходов поровну или ручными суммами;
- расчёт балансов и переводов “кто кому должен”;
- привязка расходов к маршруту, точке и времени;
- чат поездки через WebSocket;
- синхронизация точек, маршрутов, участников, расходов и сообщений в реальном времени;
- мобильный web-доступ с телефона по локальному IP;
- Android-приложение, собранное из frontend через Capacitor.

## Структура проекта

```text
travel-collab/
├── backend/                 # Go backend
│   ├── cmd/main.go          # точка входа backend
│   └── internal/
│       ├── api/             # HTTP/WebSocket handlers
│       ├── auth/            # JWT и пароли
│       ├── config/          # env-конфигурация
│       ├── db/              # подключение и миграции
│       ├── middleware/      # auth/cors middleware
│       ├── models/          # DTO и модели ответов
│       ├── repository/      # SQL-запросы
│       └── ws/              # WebSocket hub/client
├── frontend/                # Vue 3 + Vite + TypeScript + Capacitor
│   ├── src/                 # клиентское приложение
│   ├── scripts/             # mobile build scripts
│   └── android/             # Android-проект Capacitor после инициализации
├── docs/                    # тестирование и безопасность
├── scripts/                 # dev-скрипты проекта
├── sql/                     # init.sql и миграции
├── docker-compose.yml       # PostgreSQL + PostGIS
├── .env.example
└── README.md
```

## Требования

Для web-разработки:

- Windows 10/11;
- Docker Desktop;
- Go 1.23+;
- Node.js 20+;
- npm.

Для сборки Android APK дополнительно:

- Android Studio;
- Android SDK Platform-Tools;
- Android SDK Build-Tools;
- переменная `ANDROID_HOME`, например `C:\Users\valer\AppData\Local\Android\Sdk`.

## Быстрый запуск web-версии

В корне проекта:

```powershell
.\scripts\dev.ps1
```

Скрипт поднимает PostgreSQL/PostGIS через Docker Compose, ждёт готовности контейнера, запускает backend и frontend в отдельных PowerShell-окнах.

После запуска:

```text
Frontend на ПК:      http://localhost:5173
Backend API:         http://localhost:8080/api/v1
Backend health:      http://localhost:8080/api/v1/health
```

Если нужно запустить вручную:

```powershell
docker compose up -d
cd backend
$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:5433/travel_collab?sslmode=disable"
go run ./cmd/main.go
```

В отдельном терминале:

```powershell
cd frontend
npm install
npm run dev
```

## Web-версия с телефона или iPhone

Телефон должен быть подключён к Wi-Fi того же роутера, к которому подключён компьютер. У компьютера можно узнать локальный IP командой:

```cmd
ipconfig
```

Например, если IPv4 компьютера:

```text
192.168.1.166
```

то web-версия на телефоне открывается по адресу:

```text
http://192.168.1.166:5173
```

Backend при этом остаётся доступен на компьютере по порту `8080`, а frontend на телефоне обращается к нему через Vite proxy.

Если телефон не открывает frontend, проверь:

- телефон и компьютер в одной сети;
- frontend запущен через `npm run dev` с `--host 0.0.0.0`;
- Windows Firewall не блокирует Node.js или порт `5173`.

## Android APK через Capacitor

Подробная инструкция находится в файле:

```text
frontend/MOBILE.md
```

Короткий сценарий:

```powershell
cd C:\Users\valer\GolandProjects\travel-collab\frontend
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1 -BackendHost 192.168.1.166
```

После успешной сборки APK будет лежать здесь:

```text
frontend/android/app/build/outputs/apk/debug/app-debug.apk
```

APK можно перекинуть на Android-телефон и установить вручную. На телефоне не нужно устанавливать Node.js, Go, Docker или Android Studio.

## Проверка backend

Health-check endpoint:

```text
GET /api/v1/health
```

Пример:

```powershell
curl http://localhost:8080/api/v1/health
```

Ожидаемый ответ:

```json
{"status":"ok","service":"travel-collab-backend"}
```

## Проверка проекта перед коммитом

В корне проекта:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\test.ps1
```

Скрипт запускает:

- `go test ./...` для backend;
- `npm run build` для frontend.

Дополнительно ручные сценарии описаны в `docs/testing.md`.

## Основной пользовательский сценарий

1. Зарегистрировать пользователя.
2. Создать поездку.
3. Скопировать invite-код.
4. Вторым пользователем вступить в поездку по invite-коду.
5. Создать маршрут дня.
6. Добавить несколько точек на карту.
7. Изменить порядок точек через drag-and-drop.
8. Открыть маршрут в 2ГИС.
9. Добавить расход с несколькими плательщиками.
10. Проверить расчёт балансов и блок “кто кому должен”.
11. Отправить сообщение в чат.
12. Проверить синхронизацию в двух клиентах.

## Безопасность

Кратко:

- пароли хранятся в виде bcrypt-хешей;
- после входа используется JWT;
- защищённые API требуют `Authorization: Bearer ...`;
- доступ к поездке проверяется через таблицу участников;
- WebSocket-подключение проверяет JWT и принадлежность пользователя к поездке;
- секреты хранятся в `.env`, а не в Git.

Ограничения локальной версии и рекомендации для production описаны в `docs/security.md`.

## Известные ограничения

- локально используется HTTP/WS, а не HTTPS/WSS;
- Android APK в текущей версии собирается как debug APK;
- полноценная публикация в Google Play не выполнялась;
- push-уведомления не реализованы;
- для production нужно усилить CORS, добавить HTTPS/WSS, rate limiting, refresh-токены и журналирование действий.

## Материалы для диплома

Для дипломной записки удобно использовать:

- `docs/testing.md` — глава о тестировании;
- `docs/security.md` — раздел о безопасности;
- `frontend/MOBILE.md` — описание мобильной версии;
- README — краткое описание возможностей и запуска.
