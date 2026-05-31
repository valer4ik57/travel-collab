# Production checklist Travel-Collab

## Перед запуском на VPS

- [ ] Ветка `develop` актуальна.
- [ ] Создан `.env.production`.
- [ ] `APP_ENV=production`.
- [ ] `JWT_SECRET` длинный и случайный.
- [ ] `POSTGRES_PASSWORD` длинный и случайный.
- [ ] `FRONTEND_URL` указывает на публичный HTTPS-домен.
- [ ] `FRONTEND_URLS` содержит только доверенные origin.
- [ ] `TRAVEL_COLLAB_SITE_ADDRESS` равен домену сервиса.
- [ ] На домене настроена A-запись на IP сервера.
- [ ] На сервере открыты порты 80 и 443.

## После запуска

- [ ] `docker compose ps` показывает здоровые контейнеры.
- [ ] `/health` возвращает `status: ok`.
- [ ] Регистрация работает.
- [ ] Вход работает.
- [ ] Создание поездки работает.
- [ ] Invite-код работает.
- [ ] Карта и точки работают.
- [ ] Маршруты и порядок точек работают.
- [ ] Расходы и балансы работают.
- [ ] WebSocket работает между двумя устройствами.
- [ ] Android APK ходит на публичный backend, а не на localhost.

## Что сказать на защите

Публичный запуск сделан через Docker Compose. Внешняя точка входа — Caddy, который проксирует REST API и WebSocket на Go backend и отдает frontend. Backend работает с PostgreSQL/PostGIS во внутренней docker-сети. Для публичного режима добавлены строгий CORS, проверка WebSocket Origin, запрет dev JWT secret и rate limiting на регистрацию и вход.
