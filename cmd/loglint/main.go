package main

import (
	"github.com/tainj/loglint/pkg/loglint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(loglint.Analyzer)
}