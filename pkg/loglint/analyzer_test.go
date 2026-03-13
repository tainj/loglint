package loglint_test

import (
    "testing"
    "github.com/tainj/loglint/pkg/loglint"
    "golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
    testdata := analysistest.TestData()
    analysistest.Run(t, testdata, loglint.Analyzer, "example")
}