# Android APK через Capacitor

Android-версия собирается из того же Vue/Vite frontend, что и web-версия. Внутри APK работает WebView с собранными статическими файлами, а данные уходят на публичный backend Travel-Collab.

Production-сценарий:

```text
APK → https://travel-collab.ru/api/v1
APK → wss://travel-collab.ru/api/v1/ws
```

То есть сайт и APK используют одну базу, одни аккаунты, одни поездки, один чат и один backend.

## Что нужно установить

- Node.js и npm;
- Android Studio;
- Android SDK Platform-Tools;
- Android SDK Build-Tools;
- Android SDK Platform API 35 или новее;
- переменную окружения `ANDROID_HOME`.

Пример `ANDROID_HOME`:

```text
C:\Users\valer\AppData\Local\Android\Sdk
```

Проверка в PowerShell или cmd:

```cmd
adb version
echo %ANDROID_HOME%
where adb
```

## Важная настройка сервера

Capacitor APK открывает встроенный frontend с origin `https://localhost`. Поэтому production backend должен разрешать этот origin в CORS и WebSocket Origin check.

В `.env.production` на сервере в `FRONTEND_URLS` должны быть домен сайта и мобильные origins:

```env
FRONTEND_URLS=https://travel-collab.ru,https://www.travel-collab.ru,https://localhost,http://localhost,capacitor://localhost
```

После изменения `.env.production`:

```bash
cd /opt/travel-collab
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
```

Базу удалять не нужно. Не используй `down -v`.

## Сборка production APK

Из папки frontend:

```powershell
cd C:\Users\valer\GolandProjects\travel-collab\frontend
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1
```

По умолчанию скрипт собирает APK под production API:

```text
https://travel-collab.ru/api/v1
wss://travel-collab.ru/api/v1/ws
```

Готовый файл:

```text
frontend\android\app\build\outputs\apk\debug\app-debug.apk
```

## Сборка и установка сразу на телефон

1. Включи на телефоне режим разработчика.
2. Включи USB debugging.
3. Подключи телефон по USB.
4. Подтверди RSA-запрос на телефоне.
5. Проверь, что телефон виден:

```powershell
adb devices
```

Потом:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1 -Install
```

## Локальная сборка для разработки

Если нужно собрать APK под локальный backend в одной Wi-Fi сети:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1 -Mode Local -BackendHost 192.168.1.166
```

Если `-BackendHost` не указать, скрипт попробует найти локальный IP сам.

## Что проверить после установки

- запуск приложения без налезания верхней панели на системную шторку Android;
- вход по аккаунту, который уже работает на сайте;
- список поездок;
- создание поездки;
- вступление по invite-коду;
- карта и точки;
- маршруты и порядок точек;
- расходы и балансы;
- чат между APK, браузером на телефоне и браузером на компьютере.

## Частые проблемы

### APK открывается, но данные не загружаются

Проверь на сервере `FRONTEND_URLS`. Для APK нужен origin `https://localhost`.

### Чат не работает, хотя REST-запросы работают

Почти всегда причина в WebSocket Origin check. Добавь в `FRONTEND_URLS`:

```env
https://localhost,http://localhost,capacitor://localhost
```

После этого перезапусти compose без удаления базы.

### `JAVA_HOME is not set`

Скрипт пытается найти JDK Android Studio автоматически. Если не получилось:

```powershell
$env:JAVA_HOME = "C:\Program Files\Android\Android Studio\jbr"
$env:Path = "$env:JAVA_HOME\bin;$env:Path"
```

### PowerShell запрещает запуск `.ps1`

Используй запуск с `ExecutionPolicy Bypass`:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1
```
