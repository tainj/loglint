// Package example содержит тестовые кейсы для линтера loglint.
// Комментарии // want указывают ожидаемые сообщения об ошибках.
// Формат: // want "подстрока ожидаемого сообщения"
package example

import (
	"log/slog"
	"go.uber.org/zap"
)

// Глобальные переменные для тестов конкатенации
var (
	password  = "secret123"
	apiKey    = "key-abc-xyz"
	token     = "tok_12345"
	userEmail = "user@example.com"
)

// ============================================================================
// ПРАВИЛО 1: Сообщение должно начинаться со строчной буквы
// ============================================================================

func testRule1Lowercase() {
	// ❌ ОШИБКА: начинается с заглавной
	slog.Info("Starting server on port 8080")           // want "should start with lowercase"
	slog.Error("Failed to connect to database")          // want "should start with lowercase"
	zap.L().Info("User login attempt")                   // want "should start with lowercase"
	slog.Debug("Processing request")                      // want "should start with lowercase"

	// ✅ ПРАВИЛЬНО: начинается со строчной
	slog.Info("starting server on port 8080")
	slog.Error("failed to connect to database")
	zap.L().Info("user login attempt")
	slog.Debug("processing request")

	// ✅ ПРАВИЛЬНО: пустое сообщение или спецсимвол в начале (не буква)
	slog.Info("")
	slog.Info("123 numbers first")
	slog.Info("!exclamation start") // разрешено, т.к. первый символ не буква
}

// ============================================================================
// ПРАВИЛО 2: Только английский язык (кириллица и другие скрипты запрещены)
// ============================================================================

func testRule2EnglishOnly() {
	// ❌ ОШИБКА: кириллица
	slog.Info("запуск сервера")                             // want "should be in English only"
	slog.Warn("ошибка подключения к базе данных")          // want "should be in English only"
	zap.L().Debug("пользователь авторизован")              // want "should be in English only"

	// ❌ ОШИБКА: другие не-латинские скрипты
	slog.Info("服务器启动")                                   // want "should be in English only" // китайский
	slog.Info("서버 시작")                                     // want "should be in English only" // корейский

	// ✅ ПРАВИЛЬНО: только латиница, цифры, базовая пунктуация
	slog.Info("starting server")
	slog.Warn("connection timeout")
	zap.L().Debug("user authenticated successfully")
	slog.Info("api v2.1 response: status=200")

	// ✅ ПРАВИЛЬНО: слова с дефисом, слэшем, апострофом
	slog.Info("client-side rendering enabled")
	slog.Info("user's session expired")
	slog.Info("fallback/backup strategy")
}

// ============================================================================
// ПРАВИЛО 3: Без спецсимволов и эмодзи
// ============================================================================

// --- 3.1: Эмодзи запрещены ---
func testRule3aNoEmoji() {
	// ❌ ОШИБКА: эмодзи в сообщении
	slog.Info("server started! 🚀")                          // want "should not contain emoji"
	slog.Error("database connection lost 💔")               // want "should not contain emoji"
	zap.L().Warn("retry attempt #3 🔄")                     // want "should not contain emoji"
	slog.Info("success ✅")                                  // want "should not contain emoji"

	// ✅ ПРАВИЛЬНО: без эмодзи
	slog.Info("server started")
	slog.Error("database connection lost")
	zap.L().Warn("retry attempt #3")
	slog.Info("success")
}

// --- 3.2: Запрещённые спецсимволы ---
func testRule3bNoSpecialChars() {
	// ❌ ОШИБКА: запрещённые символы @#$%^&*<>[]{}|=~`"?
	slog.Info("api@host.com")                                // want "should not contain special characters"
	slog.Debug("value: $100")                               // want "should not contain special characters"
	zap.L().Info("user#12345")                              // want "should not contain special characters"
	slog.Warn("priority: [HIGH]")                           // want "should not contain special characters"
	slog.Info("result: true|false")                         // want "should not contain special characters"
	slog.Info(`quoted "value"`)                             // want "should not contain special characters"

	// ✅ ПРАВИЛЬНО: разрешённые символы .,:;!?()-/' пробел буквы цифры
	// (одиночные ! и ? в конце — разрешены, множественные — нет, см. 3.3)
	slog.Info("server started!")                             // ok: одиночный ! в конце
	slog.Info("really?")                                     // ok: одиночный ? в конце
	slog.Info("warning: something went wrong")               // ok: двоеточие разрешено
	slog.Info("status: ok, code: 200")                       // ok: запятая и двоеточие
}

// --- 3.3: Множественные знаки препинания запрещены ---
func testRule3cNoRepeatedPunctuation() {
	// ❌ ОШИБКА: повтор ! ? .
	slog.Info("connection failed!!!")                        // want "should not contain repeated punctuation"
	slog.Error("something went wrong???")                   // want "should not contain repeated punctuation"
	zap.L().Debug("loading...")                              // want "should not contain repeated punctuation"
	slog.Warn("wait...")                                     // want "should not contain repeated punctuation"

	// ✅ ПРАВИЛЬНО: одиночные знаки
	slog.Info("connection failed!")
	slog.Error("something went wrong?")
	zap.L().Debug("loading.")
}

// ============================================================================
// ПРАВИЛО 4: Без потенциально чувствительных данных
// ============================================================================

// --- 4.1: Ключевые слова в строке с конкатенацией ---
func testRule4aSensitiveConcatenation() {
	// ❌ ОШИБКА: строка содержит ключевое слово + конкатенация с переменной
	slog.Info("user password: " + password)                  // want "concatenates sensitive data"
	slog.Debug("api_key=" + apiKey)                         // want "concatenates sensitive data"
	zap.L().Info("token: " + token)                         // want "concatenates sensitive data"

	// ✅ ПРАВИЛЬНО: нет конкатенации или нет ключевого слова
	slog.Info("user authenticated successfully")
	slog.Debug("api request completed")
	zap.L().Info("token validated")                          // ok: "token" как глагол/существительное без :=+
	slog.Info("password policy updated")                     // ok: нет конкатенации, общее утверждение
}

// --- 4.2: Ключевые слова в чистом виде (без конкатенации) ---
// Примечание: чтобы избежать ложных срабатываний, проверяем только паттерны с :=
func testRule4bSensitivePatterns() {
	// ❌ ОШИБКА: паттерн "keyword:" или "keyword=" в сообщении
	slog.Info("debug: password=secret123")                   // want "may contain sensitive data"
	slog.Warn("config: api_key=prod-xyz")                   // want "may contain sensitive data"
	zap.L().Error("auth failed: token=expired")             // want "may contain sensitive data"
	slog.Info("secret: myvalue")                             // want "may contain sensitive data"

	// ✅ ПРАВИЛЬНО: слово есть, но без := после — допустимо в общем контексте
	slog.Info("password policy requires 12 characters")      // ok: общее утверждение
	slog.Info("token bucket algorithm implemented")         // ok: технический термин
	zap.L().Debug("user authentication flow started")       // ok: "authentication" не в списке
	slog.Info("credentials will be rotated")                 // ok: нет :=, общее утверждение
}

// ============================================================================
// КОМБИНИРОВАННЫЕ ТЕСТЫ: несколько правил нарушены одновременно
// ============================================================================

func testCombinedRules() {
	// ❌ Нарушает правила 1 + 3: заглавная + эмодзи
	slog.Info("Server Started 🎉")                           // want "should start with lowercase"

	// ❌ Нарушает правила 2 + 4: кириллица + чувствительное слово
	// (правило 2 сработает первым, но тест проверяет наличие хотя бы одного want)
	slog.Info("пароль: " + password)                         // want "should be in English only"

	// ❌ Нарушает правила 3 + 4: спецсимволы + чувствительные данные
	zap.L().Debug("api_key@prod=" + apiKey)                 // want "should not contain special characters"

	// ✅ Всё правильно: строчная, английский, без спецсимволов, без чувствительных
	slog.Info("user session created successfully")
	slog.Warn("rate limit approaching threshold")
	zap.L().Error("connection timeout after 30 seconds")
}

// ============================================================================
// ГРАНИЧНЫЕ СЛУЧАИ: что НЕ должно вызывать ошибок
// ============================================================================

func testEdgeCasesOk() {
	// Пустые и короткие сообщения
	slog.Info("")
	slog.Info("a")
	slog.Info("OK")                                          // ok: "OK" — устоявшаяся аббревиатура, первая буква заглавная, но можно разрешить

	// Числа, версии, коды
	slog.Info("version 2.1.0 released")
	slog.Info("http status: 404")
	slog.Info("error code: ECONNREFUSED")

	// Пути, URL (без запрещённых символов)
	slog.Info("request to /api/v1/users")
	slog.Info("redirect: https://example.com")               // :// разрешено

	// Технические термины с дефисом/слэшем
	slog.Info("client-server architecture")
	slog.Info("fallback strategy enabled")

	// Слова из списка чувствительных, но в безопасном контексте
	slog.Info("password hashing algorithm: bcrypt")          // ok: нет :=+ и конкатенации
	slog.Info("token-based authentication recommended")      // ok: общее утверждение
}