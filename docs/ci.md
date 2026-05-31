# CI для Travel-Collab

В проект добавлен GitHub Actions workflow `.github/workflows/ci.yml`.

## Что проверяется

Workflow состоит из двух независимых job-ов:

1. **Backend Go tests** — запускает `go test ./...` в папке `backend`.
2. **Frontend build** — запускает `npm ci` и `npm run build` в папке `frontend`.

Такой набор проверок закрывает две главные технические зоны:

- backend-логика расходов, security-helper-ы, rate limiting и WebSocket Origin-check;
- корректность сборки Vue/TypeScript frontend.

## Как запустить локально

Из корня проекта:

```powershell
.\scripts\test.ps1
```

Для быстрой повторной проверки без переустановки frontend-зависимостей:

```powershell
.\scripts\test.ps1 -SkipFrontendInstall
```

## Как использовать в защите

Можно сказать:

> Для проекта добавлена автоматизированная проверка: backend проходит `go test ./...`, а frontend собирается через `npm run build`. Это позволяет быстро проверить, что изменения в backend-логике и клиентской части не ломают основные технические сценарии.
