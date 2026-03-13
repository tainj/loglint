package loglint

import (
    "go/ast"
    "golang.org/x/tools/go/analysis"
    "golang.org/x/tools/go/analysis/passes/inspect"
    "golang.org/x/tools/go/ast/inspector"
)

const Doc = `Проверяет лог-сообщения на соответствие правилам:
- начинаются со строчной буквы
- только английский язык
- без спецсимволов и эмодзи
- без чувствительных данных`

var Analyzer = &analysis.Analyzer{
    Name:     "loglint",
    Doc:      Doc,
    Run:      run,
    Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
    inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
    
    // Фильтруем только узлы вызовов функций
    nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
    
    inspect.Preorder(nodeFilter, func(n ast.Node) {
        call := n.(*ast.CallExpr)
        checkLogCall(pass, call)
    })
    
    return nil, nil
}