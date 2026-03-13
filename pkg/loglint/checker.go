package loglint

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// Поддерживаемые логгеры — точные имена функций
var logFunctions = map[string]bool{
	// log/slog
	"slog.Debug": true, "slog.Info": true, "slog.Warn": true, "slog.Error": true,
	// zap (через глобальный экземпляр)
	"zap.L().Debug": true, "zap.L().Info": true, "zap.L().Warn": true, "zap.L().Error": true,
	"zap.S().Debug": true, "zap.S().Info": true, "zap.S().Warn": true, "zap.S().Error": true,
	// logger variable named log
	"log.Debug": true, "log.Info": true, "log.Warn": true, "log.Error": true,
	// stdlib log
	"log.Println": true, "log.Printf": true, "log.Print": true,
}

// Чувствительные ключевые слова
var defaultSensitive = []string{
	"password", "passwd", "pwd", "secret", "token", "api_key", "apikey",
	"credential", "private", "access_key", "session_id",
}

func checkLogCall(pass *analysis.Pass, call *ast.CallExpr) {
	funcName := getFuncName(call)
	if !isLogFunction(funcName) {
		return
	}

	// Ищем строковый аргумент (сообщение лога)
	msgArg, msgPos := findLogMessageArg(call)
	if msgArg == "" {
		return
	}

	// Правило 1: начинается со строчной
	checkLowercaseStart(pass, msgArg, msgPos)

	// Правило 2: только английский
	checkEnglishOnly(pass, msgArg, msgPos)

	// Правило 4: без чувствительных данных
	checkSensitiveData(pass, call, msgArg, msgPos)

	// Правило 3: без спецсимволов и эмодзи
	checkNoSpecialChars(pass, msgArg, msgPos)
}

// Находим первый строковый литерал в аргументах
func findLogMessageArg(call *ast.CallExpr) (string, token.Pos) {
	for _, arg := range call.Args {
		// Прямой строковый литерал
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			value := strings.Trim(lit.Value, `"`)
			return value, lit.Pos()
		}
		// Конкатенация: "msg" + var — берём строковую часть для проверки
		if bin, ok := arg.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
			if lit, ok := bin.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				value := strings.Trim(lit.Value, `"`)
				return value, lit.Pos()
			}
			if lit, ok := bin.Y.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				value := strings.Trim(lit.Value, `"`)
				return value, lit.Pos()
			}
		}
	}
	return "", token.NoPos
}

// Извлекаем имя функции из вызова (поддерживает zap.L().Info(), slog.Info() и т.д.)
func getFuncName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		// Случай: zap.L().Info() -> fn.X = CallExpr(zap.L), fn.Sel = "Info"
		if callExpr, ok := fn.X.(*ast.CallExpr); ok {
			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					// Возвращаем: "zap.L().Info"
					return ident.Name + "." + sel.Sel.Name + "()." + fn.Sel.Name
				}
			}
		}
		// Простой случай: slog.Info -> fn.X = Ident("slog"), fn.Sel = "Info"
		if ident, ok := fn.X.(*ast.Ident); ok {
			return ident.Name + "." + fn.Sel.Name
		}
	case *ast.Ident:
		return fn.Name
	}
	return ""
}

// Проверка функции (относится ли она к логированию)
func isLogFunction(funcName string) bool {
	// Проверяем имя метода (универсально для logger.*, myLog.*, etc.)
	// Разбиваем "logger.Info" -> ["logger", "Info"]
	parts := strings.Split(funcName, ".")
	if len(parts) >= 2 {
		method := parts[len(parts)-1] // берём последнее: "Info", "Error", etc.
		
		// Если метод — лог-метод, разрешаем любую переменную перед ним
		allowedMethods := []string{
			"Info", "Error", "Warn", "Debug",
			"Infof", "Errorf", "Warnf", "Debugf",
			"Print", "Printf", "Println",
		}
		for _, m := range allowedMethods {
			if method == m {
				return true
			}
		}
	}
	
	// 2. Фоллбэк: точное совпадение из словаря (для slog.*, zap.L().*)
	if logFunctions[funcName] {
		return true
	}
	
	return false
}

// Правило 1: первая буква — строчная
func checkLowercaseStart(pass *analysis.Pass, msg string, pos token.Pos) {
	if msg == "" {
		return
	}
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	r, _ := utf8.DecodeRuneInString(msg)
	// Флагим только если первая руна — ЗАГЛАВНАЯ буква (не цифра, не спецсимвол)
	if unicode.IsLetter(r) && unicode.IsUpper(r) {
		pass.Reportf(pos, "log message should start with lowercase letter: %q", msg)
	}
}

// Вспомогательная: разрешённые символы для английского текста
func isAllowedEnglishSymbol(r rune) bool {
	// Латиница, цифры, пробел, базовая пунктуация для логов
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == ' ' || r == '.' || r == ',' ||
		r == ':' || r == ';' || r == '!' || r == '?' ||
		r == '-' || r == '(' || r == ')' || r == '/' || r == '\'' ||
		r == '"' || r == '_' || r == '=' || r == '@' || r == '&'
}

// Правило 2: только английский (кириллица и другие не-латинские скрипты запрещены)
func checkEnglishOnly(pass *analysis.Pass, msg string, pos token.Pos) {
	for _, r := range msg {
		// Если символ не в разрешённом наборе — проверяем, не запрещённый ли это скрипт
		if !isAllowedEnglishSymbol(r) {
			// Кириллица
			if (r >= 0x0400 && r <= 0x04FF) || (r >= 0x0500 && r <= 0x052F) {
				pass.Reportf(pos, "log message should be in English only: %q", msg)
				return
			}
			// Китайские иероглифы
			if r >= 0x4E00 && r <= 0x9FFF {
				pass.Reportf(pos, "log message should be in English only: %q", msg)
				return
			}
			// Корейский
			if (r >= 0xAC00 && r <= 0xD7AF) || (r >= 0x1100 && r <= 0x11FF) {
				pass.Reportf(pos, "log message should be in English only: %q", msg)
				return
			}
			// Арабский, иврит и др. — можно добавить при необходимости
		}
	}
}

// Правило 3: без спецсимволов и эмодзи
func checkNoSpecialChars(pass *analysis.Pass, msg string, pos token.Pos) {
	// 1. Эмодзи — запрещены всегда
	for _, r := range msg {
		if isEmoji(r) {
			pass.Reportf(pos, "log message should not contain emoji: %q", msg)
			return
		}
	}

	// 2. Запрещённые спецсимволы (кроме разрешённой пунктуации)
	// Разрешено: .,:;!?()-/'"= @& пробел буквы цифры
	// Запрещено: @#$%^&*+=<>[]{}|~` (но @ и & оставили разрешёнными для email и т.п.)
	forbidden := "#$%^*+<>[]{}|~`\\"
	for _, r := range msg {
		if strings.ContainsRune(forbidden, r) {
			pass.Reportf(pos, "log message should not contain special characters: %q", msg)
			return
		}
	}

	// 3. Повторы пунктуации: !!! ??? ... (3 и более подряд)
	if regexp.MustCompile(`[!?]{2,}|\.{3,}`).MatchString(msg) {
		pass.Reportf(pos, "log message should not contain repeated punctuation: %q", msg)
	}
}

// Вспомогательная: проверка на эмодзи по диапазонам Unicode
func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F1E0 && r <= 0x1F1FF) // Flags
}

// Правило 4: чувствительные данные
func checkSensitiveData(pass *analysis.Pass, call *ast.CallExpr, msg string, pos token.Pos) {
	msgLower := strings.ToLower(msg)

	// 4.1: Проверка конкатенации с переменными: "password: " + pwd
	for _, arg := range call.Args {
		if bin, ok := arg.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
			// Проверяем оба операнда на наличие строкового литерала с ключевым словом
			for _, expr := range []ast.Expr{bin.X, bin.Y} {
				if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					litValue := strings.ToLower(strings.Trim(lit.Value, `"`))
					for _, kw := range defaultSensitive {
						if strings.Contains(litValue, kw) {
							pass.Reportf(pos, "log message concatenates sensitive data: %q", kw)
							return
						}
					}
				}
			}
		}
	}

	// 4.2: Проверка паттернов вида "password=...", "api_key: ..."
	// ИЗМЕНЕНО: требуем := после ключевого слова, чтобы избежать ложных срабатываний
	// на "password policy", "token bucket" и т.п.
	for _, kw := range defaultSensitive {
		// Ищем паттерны: keyword= или keyword: (с пробелом или без)
		patterns := []string{kw + "=", kw + ": ", kw + ":"}
		for _, p := range patterns {
			if strings.Contains(msgLower, p) {
				pass.Reportf(pos, "log message may contain sensitive data: %q", kw)
				return
			}
		}
	}
}