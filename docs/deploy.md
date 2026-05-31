# Production-деплой Travel-Collab

Документ описывает первый production-ready запуск проекта через Docker Compose. Схема рассчитана на небольшой VPS: Caddy принимает HTTP/HTTPS, отдает frontend и проксирует REST API/WebSocket на Go backend, backend работает с PostgreSQL/PostGIS.

## Состав контейнеров

- `postgres` — PostgreSQL/PostGIS, данные хранятся в Docker volume `travel_collab_pgdata_prod`.
- `backend` — Go REST API + WebSocket Hub.
- `frontend` — собранный Vue/Vite frontend в nginx.
- `caddy` — reverse proxy, HTTPS/WSS и входная точка сервиса.

## Быстрая локальная проверка production-сборки

Из корня проекта в PowerShell:

```powershell
.\scripts\prod-up.ps1
```

Скрипт сам создаст локальный `.env.production`, если файла нет. После сборки открой:

```text
http://localhost
```

Остановка:

```powershell
.\scripts\prod-down.ps1
```

Если порт 80 занят, измени `HTTP_PORT` в `.env.production`, например:

```env
HTTP_PORT=8088
```

Тогда сервис будет доступен по `http://localhost:8088`, но для корректного CORS нужно также добавить этот адрес в `FRONTEND_URL` и `FRONTEND_URLS`.

## Подготовка `.env.production` для сервера

Создать файл можно командой:

```powershell
.\scripts\init-production-env.ps1 -Domain travel-collab.example.ru
```

Затем открыть `.env.production` и проверить значения:

```env
APP_ENV=production
TRAVEL_COLLAB_SITE_ADDRESS=travel-collab.example.ru
FRONTEND_URL=https://travel-collab.example.ru
FRONTEND_URLS=https://travel-collab.example.ru
DATABASE_URL=postgres://travel_collab:<password>@postgres:5432/travel_collab?sslmode=disable
JWT_SECRET=<long-random-secret>
```

В production backend не стартует, если `JWT_SECRET` короткий, содержит шаблонные значения или dev-секрет.

## Запуск на сервере

На VPS должны быть установлены Docker и Docker Compose plugin.

```bash
cd /opt/travel-collab
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
```

Проверка контейнеров:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

Логи backend:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml logs -f backend
```

Логи Caddy:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml logs -f caddy
```

## DNS и HTTPS

Для домена нужно создать A-запись на публичный IP сервера. После этого Caddy автоматически выпустит сертификат. REST API будет работать через HTTPS, WebSocket — через WSS.

Входные порты на сервере должны быть открыты:

- `80/tcp` — HTTP и выпуск сертификата;
- `443/tcp` — HTTPS/WSS.

PostgreSQL наружу не публикуется. Backend тоже не публикуется напрямую, он доступен только через Caddy.

## Проверка после запуска

1. Открыть главную страницу.
2. Зарегистрировать нового пользователя.
3. Создать поездку.
4. Создать маршрут и добавить точки.
5. Проверить линию между точками и изменение порядка.
6. Добавить расход с несколькими плательщиками или ручным делением.
7. Проверить баланс и взаиморасчеты.
8. Открыть ту же поездку во втором браузере/устройстве и проверить WebSocket-обновления.

## Что уже усилено для публичного режима

- CORS в production пропускает только явно указанные `FRONTEND_URL`/`FRONTEND_URLS`.
- WebSocket проверяет `Origin`, а не принимает любые подключения.
- В production запрещен слабый или dev `JWT_SECRET`.
- На `login` и `register` добавлен in-memory rate limit.
- Размер входящего запроса ограничен через `MAX_REQUEST_BODY_BYTES`.
- PostgreSQL не публикуется наружу.

### Частая ошибка: backend не может распарсить DATABASE_URL

Если в логах backend есть ошибка вида `failed to parse as URL` или `invalid port ... after host`, почти всегда причина в том, что пароль PostgreSQL содержит символы `@`, `:`, `/`, `?`, `#`, `&`, `+` или `=` и при прямой вставке в `DATABASE_URL` ломает URL.

Для простоты используйте пароль БД только из латинских букв и цифр. Скрипт `scripts/init-production-env.ps1` уже генерирует именно такой пароль. Если ошибка возникла на первом запуске, можно безопасно пересоздать локальную production-базу:

```powershell
.\scripts\prod-down.ps1 -Volumes
Remove-Item .env.production
.\scripts\prod-up.ps1
```

Флаг `-Volumes` удаляет volume PostgreSQL, поэтому используйте его только до появления важных данных или после создания бэкапа.
