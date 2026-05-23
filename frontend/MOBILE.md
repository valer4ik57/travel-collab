# Android-приложение Travel-Collab через Capacitor

Android-версия собирается из того же Vue/Vite frontend, что и web-версия. Внутри APK работает WebView, а данные загружаются с Go backend, запущенного на компьютере в локальной сети.

## Что нужно установить на компьютер

- Node.js и npm;
- Android Studio;
- Android SDK Platform-Tools;
- Android SDK Build-Tools;
- Android SDK Platform API 35;
- переменную окружения `ANDROID_HOME`.

Пример `ANDROID_HOME`:

```text
C:\Users\valer\AppData\Local\Android\Sdk
```

Проверка:

```cmd
adb version
echo %ANDROID_HOME%
where adb
```

На телефон ничего дополнительного ставить не нужно, кроме готового APK.

## Подготовка сети

Компьютер и телефон должны быть в одной локальной сети. Компьютер может быть подключён к роутеру по Ethernet, а телефон — по Wi-Fi.

Узнать IP компьютера:

```cmd
ipconfig
```

Нужен IPv4 активного адаптера, например:

```text
192.168.1.166
```

## Сборка APK

Сначала запусти backend и frontend обычным способом из корня проекта:

```powershell
.\scripts\dev.ps1
```

Потом в отдельном PowerShell:

```powershell
cd C:\Users\valer\GolandProjects\travel-collab\frontend
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1 -BackendHost 192.168.1.166
```

Скрипт выполнит:

1. установку npm-зависимостей;
2. сборку frontend через Vite;
3. синхронизацию Capacitor;
4. сборку debug APK через Gradle.

Готовый APK:

```text
frontend/android/app/build/outputs/apk/debug/app-debug.apk
```

## Установка APK на телефон

1. Перекинь `app-debug.apk` на телефон.
2. Открой APK через файловый менеджер, браузер или мессенджер.
3. Разреши установку из неизвестных источников для выбранного приложения.
4. Установи Travel-Collab.

## Что проверить после установки

- вход по email/паролю;
- список поездок;
- открытие поездки;
- карта и точки;
- маршруты и drag-and-drop порядка точек;
- открытие маршрута в 2ГИС;
- расходы и расчёт балансов;
- чат;
- WebSocket-синхронизация;
- отсутствие налезания интерфейса на системную шторку Android.

## Частые проблемы

### APK открывается, но данные не загружаются

Проверь:

- backend запущен;
- APK собран с правильным `-BackendHost`;
- телефон и компьютер находятся в одной сети;
- Windows Firewall не блокирует порт `8080`.

### Телефон не открывает web-версию

Открывай frontend, а не backend:

```text
http://192.168.1.166:5173
```

Backend напрямую проверяется через:

```text
http://192.168.1.166:8080/api/v1/health
```

### `JAVA_HOME is not set`

Скрипт пытается найти JDK Android Studio автоматически. Если не получилось:

```powershell
$env:JAVA_HOME = "C:\Program Files\Android\Android Studio\jbr"
$env:Path = "$env:JAVA_HOME\bin;$env:Path"
```

### Не установлен Android SDK Platform 35

Открой Android Studio:

```text
Tools → SDK Manager → SDK Platforms
```

Установи обычный:

```text
Android 15.0 / API 35
```

Не обязательно выбирать `35-ext...`.

### PowerShell запрещает запуск `.ps1`

Используй запуск с `ExecutionPolicy Bypass`:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1 -BackendHost 192.168.1.166
```
