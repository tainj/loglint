package main

import "log/slog"

func main() {
	// Все эти строки должны проходить без ошибок
	slog.Info("starting server")
	slog.Info("server started")
	slog.Info("status=200")
	slog.Info("password policy updated")      // ok: безопасный контекст
	slog.Info("token bucket algorithm")       // ok: технический термин
	slog.Info("connection failed!")           // ok: одиночный ! разрешён
}