package example

import (
    "log"
    "go.uber.org/zap"
    "log/slog"
)

func test() {
    // ❌ Ошибки
    log.Println("Starting server")           // want "should start with lowercase"
    slog.Error("Failed to connect")          // want "should start with lowercase"
    log.Println("запуск сервера")            // want "should be in English only"
    log.Println("server started! 🚀")        // want "should not contain emoji"
    log.Println("connection failed!!!")      // want "should not contain repeated punctuation"
    log.Println("user password: " + pwd)     // want "may contain sensitive data"
    
    // ✅ Правильно
    log.Println("starting server")
    slog.Error("failed to connect")
    log.Println("server started")
    log.Println("connection failed")
    log.Println("user authenticated")
}