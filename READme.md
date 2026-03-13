# loglint 🔍

Линтер для анализа лог-записей в Go-проектах. Проверяет сообщения на соответствие стандартам качества и безопасности.

## ✨ Возможности

Линтер проверяет 4 правила для вызовов `slog` и `zap`:

| Правило | Описание | Пример ❌ | Пример ✅ |
|---------|----------|-----------|-----------|
| **Lowercase start** | Сообщение должно начинаться со строчной буквы | `slog.Info("Starting")` | `slog.Info("starting")` |
| **English only** | Только латиница, цифры и базовая пунктуация | `slog.Info("запуск")` | `slog.Info("starting")` |
| **No emoji/special chars** | Без эмодзи и запрещённых символов `#$%^&*+<>[]{}\|~\`` | `slog.Info("ok 🚀")` | `slog.Info("ok")` |
| **No sensitive data** | Без паролей, токенов, API-ключей | `slog.Info("pwd: " + pass)` | `slog.Info("auth completed")` |

### Поддерживаемые логгеры

- **`log/slog`**: `slog.Info()`, `slog.Error()`, `slog.Warn()`, `slog.Debug()` (+f версии)
- **`go.uber.org/zap`**: 
  - Через переменную: `logger.Info()`, `logger.Errorf()`, etc.
  - Глобальный: `zap.L().Info()`, `zap.S().Error()`, etc.

---

## 🚀 Установка и использование

### Standalone (рекомендуется)

```bash
# Установка
go install github.com/tainj/loglint/cmd/loglint@latest

# Запуск в текущем проекте
loglint ./...

# Запуск в конкретной папке
loglint ./pkg/...
```

### Интеграция с golangci-lint

После публикации модуля добавьте в `.golangci.yml`:

```yaml
linters-settings:
  custom:
    loglint:
      type: "module"
      path: "github.com/tainj/loglint/pkg/loglint"
linters:
  enable:
    - loglint
```

> ⚠️ Для локальной разработки используйте standalone-режим. Интеграция с golangci-lint требует публикации модуля.

---

## 🧪 Тестирование

### Unit-тесты

```bash
# Запустить все тесты
go test ./pkg/loglint/... -v

# Запустить с покрытием
go test ./pkg/loglint/... -cover
```

### Демо-проекты

В репозитории есть примеры для проверки:

```bash
# Код с нарушениями (должен показать ошибки)
cd demo/violations
../../loglint ./...

# Чистый код (должен пройти без вывода)
cd demo/clean
../../loglint ./...
```

**Ожидаемый вывод для `violations`:**
```
main.go:7:12: log message should start with lowercase letter: "Starting server"
main.go:8:12: log message should be in English only: "запуск сервера"
main.go:9:12: log message should not contain emoji: "server started 🚀"
...
```

---

## 🔄 CI/CD

Настроен GitHub Actions с проверками:

- `test` — unit-тесты анализатора
- `demo-violations` — детектирование нарушений
- `demo-clean` — чистый код проходит без ошибок

[![CI](https://github.com/tainj/loglint/actions/workflows/ci.yml/badge.svg)](https://github.com/tainj/loglint/actions/workflows/ci.yml)

---

## 📁 Структура проекта

```
.
├── cmd/
│   └── loglint/
│       └── main.go          # точка входа standalone CLI
├── pkg/
│   └── loglint/
│       ├── analyzer.go      # ядро анализатора (правила)
│       ├── analyzer_test.go # unit-тесты
│       └── testdata/        # тестовые данные для analysistest
├── demo/
│   ├── violations/          # пример кода с нарушениями
│   └── clean/               # пример чистого кода
├── .github/workflows/
│   └── ci.yml               # конфигурация GitHub Actions
├── go.mod
├── go.sum
└── README.md
```
