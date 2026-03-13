package loglint

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"golang.org/x/text/unicode/rangetable"
	"golang.org/x/tools/go/analysis"
)

// Поддерживаемые логгеры
var logFunctions = map[string]bool{
    // log/slog
    "slog.Debug": true, "slog.Info": true, "slog.Warn": true, "slog.Error": true,
    // zap
    "zap.Debug": true, "zap.Info": true, "zap.Warn": true, "zap.Error": true,
    "logger.Debug": true, "logger.Info": true, "logger.Warn": true, "logger.Error": true,
    // stdlib log
    "log.Println": true, "log.Printf": true, "log.Print": true,
}

// Чувствительные ключевые слова
var defaultSensitive = []string{
    "password", "passwd", "pwd", "secret", "token", "api_key", "apikey",
    "auth", "credential", "private", "access_key", "session_id",
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
    
    // Правило 3: без спецсимволов и эмодзи
    checkNoSpecialChars(pass, msgArg, msgPos)
    
    // Правило 4: без чувствительных данных
    checkSensitiveData(pass, call, msgArg, msgPos)
}

// Находим первый строковый литерал в аргументах
func findLogMessageArg(call *ast.CallExpr) (string, token.Pos) {
    for _, arg := range call.Args {
        if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
            // Убираем кавычки
            value := strings.Trim(lit.Value, `"`)
            return value, lit.Pos()
        }
        // Обработка конкатенации: "msg" + var
        if bin, ok := arg.(*ast.BinaryExpr); ok {
            if lit, ok := bin.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
                value := strings.Trim(lit.Value, `"`)
                return value, lit.Pos()
            }
        }
    }
    return "", token.NoPos
}

// Извлекаем имя функции из вызова
func getFuncName(call *ast.CallExpr) string {
    switch fn := call.Fun.(type) {
    case *ast.SelectorExpr:
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
    for logFuncName := range logFunctions {
        if strings.HasSuffix(funcName, logFuncName) || logFuncName == funcName {
            return true
        }
    }
    return false
}

// Правило 1: первая буква — строчная
func checkLowercaseStart(pass *analysis.Pass, msg string, pos token.Pos) {
	if msg == "" {
        return
    }
    r, _ := utf8.DecodeLastRuneInString(msg)
    if unicode.IsUpper(r) {
        pass.Reportf(pos, "log message should start with lowercase letter: %q", msg)
    }
}

// Правило 2: только английский (проверка через unicode ranges)
func checkEnglishOnly(pass *analysis.Pass, msg string, pos token.Pos) {
    // Разрешённые диапазоны: базовая латиница + знаки препинания + пробелы
    allowed := rangetable.Merge(
        unicode.BasicLatin,      // A-Z, a-z, 0-9, punctuation
        unicode.Latin1Supplement, // расширенная латиница (опционально)
    )
    
    for _, r := range msg {
        if !unicode.In(r, allowed) && !unicode.IsSpace(r) {
            // Если символ не в разрешённых диапазонах — вероятно, не английский
            if unicode.Is(unicode.Cyrillic, r) || unicode.Is(unicode.Han, r) {
                pass.Reportf(pos, "log message should be in English only: %q", msg)
                return
            }
        }
    }
}

// Правило 3: без спецсимволов и эмодзи
func checkNoSpecialChars(pass *analysis.Pass, msg string, pos token.Pos) {
    // Разрешаем: буквы, цифры, пробелы, базовая пунктуация .,:;!?-()
    allowed := regexp.MustCompile(`^[a-zA-Z0-9\s\.\,\:\;\!\?\-\(\)\/\\']+$`)
    
    if !allowed.MatchString(msg) {
        // Проверяем наличие эмодзи (диапазоны Unicode)
        for _, r := range msg {
            if unicode.In(r, unicode.Emoji) {
                pass.Reportf(pos, "log message should not contain emoji: %q", msg)
                return
            }
            // Спецсимволы: множественные !?..., символы @#$%^&*<>[]{}|=~`
            if strings.ContainsAny(string(r), "!@#$%^&*<>[]{}|=~`") {
                // Разрешаем одиночные !? в конце
                if !strings.HasSuffix(msg, "!") && !strings.HasSuffix(msg, "?") {
                    pass.Reportf(pos, "log message should not contain special characters: %q", msg)
                    return
                }
            }
        }
    }
    
    // Проверка на множественные знаки препинания
    if regexp.MustCompile(`[!?]{2,}|\.{3,}`).MatchString(msg) {
        pass.Reportf(pos, "log message should not contain repeated punctuation: %q", msg)
    }
}

// Правило 4: чувствительные данные
func checkSensitiveData(pass *analysis.Pass, call *ast.CallExpr, msg string, pos token.Pos) {
    msgLower := strings.ToLower(msg)
    
    // Проверка по ключевым словам
    for _, keyword := range defaultSensitive {
        if strings.Contains(msgLower, keyword) {
            pass.Reportf(pos, "log message may contain sensitive  %q", keyword)
            return
        }
    }
    
    // Проверка конкатенации с переменными
    for _, arg := range call.Args {
        if bin, ok := arg.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
            // "text" + variable — потенциально чувствительно
            if lit, ok := bin.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
                litValue := strings.Trim(lit.Value, `"`)
                for _, keyword := range defaultSensitive {
                    if strings.Contains(strings.ToLower(litValue), keyword) {
                        pass.Reportf(pos, "log message concatenates sensitive data: %q", keyword)
                        return
                    }
                }
            }
        }
    }
}