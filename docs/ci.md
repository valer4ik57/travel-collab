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
