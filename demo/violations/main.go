package main

import "log/slog"

func main() {
	// Все эти строки должны детектироваться как нарушения
	slog.Info("Starting server")              // ❌ заглавная буква
	slog.Info("запуск сервера")               // ❌ кириллица
	slog.Info("server started 🚀")            // ❌ эмодзи
	slog.Info("value: $100")                  // ❌ спецсимвол $
	slog.Info("failed!!!")                    // ❌ повтор пунктуации
	slog.Info("password: " + "secret123")     // ❌ конкатенация чувствительных
	slog.Info("debug: password=123")          // ❌ паттерн password=
}