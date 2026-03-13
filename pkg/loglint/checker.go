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

// Вспомогательная функция: проверка на числа и латиницу
func isAllowedSymbol(r rune) bool {
    return (r >= 'a' && r <= 'z') ||
                     (r >= 'A' && r <= 'Z') ||
                     (r >= '0' && r <= '9') || 
                     r == ' ' || r == '.' || r == ',' || 
                     r == ':' || r == ';' || r == '!' || r == '?' || 
                     r == '-' || r == '(' || r == ')' || r == '/' || r == '\''
}

// Правило 2: только английский (проверка через unicode ranges)
func checkEnglishOnly(pass *analysis.Pass, msg string, pos token.Pos) {
    // Разрешённые диапазоны: базовая латиница + знаки препинания + пробелы
    for _, symbol := range msg {
        if !isAllowedSymbol(symbol) {
            // Явно проверяем "запрещённые" скрипты
            if (symbol >= 0x0400 && symbol <= 0x04FF) { // кириллица
                pass.Reportf(pos, "log message should be in English only: %q", msg)
                return
            }
        }
    }
}

// Правило 3: без спецсимволов и эмодзи
// Правило 3: без спецсимволов и эмодзи
func checkNoSpecialChars(pass *analysis.Pass, msg string, pos token.Pos) {
    // 1. Проверка на эмодзи по диапазонам Unicode
    for _, r := range msg {
        // Основные диапазоны эмодзи
        if isEmoji(r) {
            pass.Reportf(pos, "log message should not contain emoji: %q", msg)
            return
        }
    }
    
    // 2. Проверка на запрещённые спецсимволы
    // Разрешаем только: буквы, цифры, пробел, . , : ; - ( ) / '
    forbidden := "!@#$%^&*<>[]{}|=~`\"+\\?"
    for _, r := range msg {
        if strings.ContainsRune(forbidden, r) {
            pass.Reportf(pos, "log message should not contain special characters: %q", msg)
            return
        }
    }
    
    // 3. Проверка на множественные знаки препинания: !!!, ???, ...
    if regexp.MustCompile(`[!?]{2,}|\.{3,}`).MatchString(msg) {
        pass.Reportf(pos, "log message should not contain repeated punctuation: %q", msg)
    }
}

// Вспомогательная функция: проверка на эмодзи по диапазонам
func isEmoji(r rune) bool {
    return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
           (r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
           (r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
           (r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols
           (r >= 0x2600 && r <= 0x26FF) ||   // Misc symbols
           (r >= 0x2700 && r <= 0x27BF) ||   // Dingbats
           (r >= 0xFE00 && r <= 0xFE0F) ||   // Variation Selectors
           (r >= 0x1F1E0 && r <= 0x1F1FF)    // Flags
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