package loglint

import (
    "testing"
    "golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
    // "./src/example" — относительный путь от testdata/
    analysistest.Run(t, analysistest.TestData(), Analyzer, "./src/example")
}