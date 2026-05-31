# Трассировка требований Travel-Collab

Трассировка требований показывает, где каждое важное требование реализовано в проекте и как оно проверяется. Этот документ нужен не для работы приложения, а для защиты и финальной сверки ВКР: он связывает текст требований, код, базу данных и тестирование.

## Таблица трассировки

| Требование | Где реализовано | Как проверяется | Что показывать/говорить на защите |
|---|---|---|---|
| Регистрация пользователя | `backend/internal/api/auth_handlers.go`, `backend/internal/auth/password.go`, таблица `users` | Ручной сценарий регистрации; backend-тесты паролей можно расширить при необходимости | Пароль не хранится открытым текстом, backend сохраняет bcrypt-хеш |
| Вход по email и паролю | `auth_handlers.go`, `backend/internal/auth/jwt.go` | Ручной вход; негативный сценарий с неверным паролем | После успешного входа backend выдает JWT |
| Защита REST API через JWT | `backend/internal/middleware/auth.go`, `server.go` | Запрос к защищенному endpoint-у без JWT должен вернуть 401 | JWT отвечает на вопрос “кто пользователь”, но не определяет права в поездке |
| Создание поездки | `trip_handlers.go`, `repository/trips.go`, таблица `trips` | Ручной сценарий создания поездки | Создатель поездки становится owner |
| Присоединение по invite-коду | `trip_handlers.go`, `repository/trips.go`, `trips.invite_code`, `trip_members` | Ручной сценарий второго пользователя | Invite-код добавляет пользователя в конкретную поездку |
| Роли owner/editor/viewer | `trip_members.role`, `ensureTripRole`, `repository/trips.go` | Ручные негативные сценарии viewer/editor; backend-тесты можно расширять | Права задаются не глобально, а внутри конкретной поездки |
| Запрет потери последнего owner | `handleUpdateMemberRole`, `handleRemoveMember`, `handleLeaveTrip`, repository-методы подсчета владельцев | Негативный сценарий: нельзя понизить/удалить последнего owner | Поездка не должна оставаться без владельца |
| Маршруты поездки | `trip_routes`, `route_handlers.go`, `repository/routes.go` | Создание/редактирование/удаление маршрута | Маршрут — план посещения, а не автомобильная навигация |
| Порядок точек маршрута | `locations.route_id`, `locations.route_order`, `handleReorderRouteLocations` | Drag-and-drop точек; ручная проверка линии на карте | Порядок хранится на backend, линия показывает пользовательскую последовательность точек |
| Точки интереса на карте | `locations`, PostGIS `GEOMETRY(Point, 4326)`, `repository/locations.go`, frontend Leaflet | Добавление/редактирование точки; отображение маркера | Координаты хранятся как геоданные, а не только как текст |
| Связь расхода с маршрутом или точкой | `expenses.route_id`, `expenses.location_id`, `normalizeAndValidateExpense` | Негативные сценарии чужой точки/маршрута; ручное создание расхода по точке | Расход может относиться не только к поездке, но и к конкретному контексту маршрута |
| Несколько плательщиков расхода | `expenses.payments`, `ExpensePayment`, `normalizeAndValidateExpense` | Ручной сценарий “несколько плательщиков”; unit-тесты totals | Плательщики и участники деления могут не совпадать |
| Ручное распределение расходов | `expenses.shares`, `ExpenseShare`, `split_mode=manual` | Ручной сценарий с ручными суммами долей | В интерфейсе оставлены понятные сценарии: поровну и ручные суммы |
| Проверка финансовой целостности | `paymentsTotalCents`, `sharesTotalCents`, `normalizeAndValidateExpense` | Unit-тесты расходов; негативные сценарии с неверной суммой оплат/долей | Backend отклоняет некорректный расход, frontend не является источником истины |
| Расчет балансов | `repository/expenses.go`, `GetExpensesSummary` | Unit-тесты `effectiveShares`, `effectivePayments`, ручная проверка списка балансов | Баланс = оплачено участником − его доля расходов |
| Расчет взаиморасчетов | `buildSettlements` | Unit-тесты `BuildSettlements...` | Должники закрывают задолженности перед кредиторами до нуля |
| Чат поездки | `messages`, `message_handlers.go`, `repository/messages.go` | Отправка и загрузка истории сообщений | Сообщения привязаны к `trip_id`, поэтому обсуждения разных поездок не смешиваются |
| WebSocket-обновления | `/api/v1/ws/{trip_id}`, `websocket_handler.go`, `ws/hub.go`, `ws/client.go` | Проверка двумя клиентами; unit-тест Origin-check | Hub хранит комнаты по `trip_id` и рассылает события только участникам поездки |
| WebSocket-доступ только участникам | `handleWebSocket`, `ParseToken`, `IsTripMember` | Негативный сценарий подключения к чужой поездке | Даже real-time соединение проходит проверку JWT и членства в поездке |
| CORS и Origin в production | `config.Config.IsOriginAllowed`, `middleware/cors.go`, `websocket_handler.go` | Unit-тесты config и WebSocket Origin | В production разрешаются только явно заданные frontend-домены |
| Rate limiting авторизации | `middleware/rate_limit.go`, `server.go` | Unit-тесты rate limiter; ручная проверка частых запросов | Login/register защищены от примитивного перебора и спама запросами |
| Production JWT secret | `config.go`, `isWeakJWTSecret` | Unit-тесты слабого секрета; запуск production без секрета должен падать | В production нельзя стартовать с dev-секретом |
| Ограничение размера запросов | `middleware.RequestSize(s.cfg.MaxRequestBodyBytes)` в `server.go` | Ручная/автоматическая проверка при необходимости | Backend ограничивает размер входящего тела запроса |
| Web-версия | `frontend`, Vue 3, TypeScript, Vite | `npm run build`, ручная проверка UI | Один frontend используется для браузера и мобильной версии |
| Android APK через Capacitor | `frontend/capacitor.config.ts`, Capacitor-зависимости | Ручная debug-сборка и запуск APK | Android-сценарий построен на том же frontend-коде |
| Публичное развертывание | `docker-compose.prod.yml`, `Caddyfile`, `backend/Dockerfile`, `frontend/Dockerfile`, `docs/deploy.md` | `scripts/prod-up.ps1`, проверка healthcheck и пользовательских сценариев | Проект можно поднять как публичный демо-сервис с reverse proxy |
| Автоматическая проверка сборки | `.github/workflows/ci.yml`, `scripts/test.ps1` | GitHub Actions, локальный `scripts/test.ps1` | Backend-тесты и frontend-сборка проверяются автоматически |

## Короткая формула для ответа комиссии

Если спрашивают “где это реализовано?”, отвечать по цепочке:

```text
требование → таблица/endpoint → backend-проверка → тест/ручной сценарий → результат в интерфейсе
```

Например, для расходов:

```text
расширенный расход → таблица expenses(payments, shares, split_mode, route_id, location_id)
→ normalizeAndValidateExpense проверяет суммы, участников и связи с поездкой
→ GetExpensesSummary считает балансы и взаиморасчеты
→ тесты expenses_test.go и ручной сценарий с несколькими плательщиками
```

Для доступа:

```text
JWT → user_id → trip_members(trip_id, user_id) → role → разрешение/запрет действия
```
