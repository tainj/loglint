// Package example — тестовые кейсы для loglint.
// Ожидания указываются в комментариях после строк кода.
package example

import (
	"log/slog"
	"go.uber.org/zap"
)

var password = "secret"

func test() {
	// Правило 1: строчная буква в начале
	slog.Info("Starting server")              // want "should start with lowercase"
	zap.L().Error("Failed connection")        // want "should start with lowercase"
	
	slog.Info("starting server")              // ok
	zap.L().Error("failed connection")        // ok
	
	// Правило 2: только английский
	slog.Info("запуск сервера")               // want "should be in English only"
	slog.Info("starting server")              // ok
	slog.Info("api v2.1 response: status=200")  // ok
	
	// Правило 3: эмодзи и спецсимволы
	slog.Info("server started 🚀")            // want "should not contain emoji"
	slog.Info("value: $100")                  // want "should not contain special characters"
	slog.Info("user#123")                     // want "should not contain special characters"
	slog.Info("failed!!!")                    // want "should not contain repeated punctuation"
	
	slog.Info("connection failed!")           // ok
	slog.Info("really?")                      // ok
	slog.Info("status=200")                   // ok
	slog.Info("path: /api/v1")                // ok
	
	// Правило 4: чувствительные данные
	// Примечание: сообщения об ошибках должны совпадать с тем, что в Reportf()
	slog.Info("password: " + password)        // want "concatenates sensitive"
	slog.Info("debug: password=secret123")    // want "may contain sensitive"
	
	slog.Info("password policy updated")      // ok
	slog.Info("token bucket algorithm")       // ok
}