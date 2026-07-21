package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/cmd/staticlint/osexitcheck"
	"github.com/kisielk/errcheck/errcheck"
	gosec "github.com/securego/gosec/v2/analyzers"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

const ignoreFile = `staticcheck_ignore.json`

type StaticChecks []string

func main() {
	// возвращает полный путь к исполняемому файлу текущей программы.
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), ignoreFile))
	if err != nil {
		panic(err)
	}
	var ignoreChecks StaticChecks
	if err = json.Unmarshal(data, &ignoreChecks); err != nil {
		panic(err)
	}

	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		nilness.Analyzer,
		errcheck.Analyzer,
		bodyclose.Analyzer,
		osexitcheck.Analyzer,
	}

	// Анализ потенциальных уязвимостей безопасности
	gosecAll := gosec.BuildDefaultAnalyzers()
	// Игнорирование конфликтных
	brokenSSA := map[string]bool{
		"G115": true, // integer overflow conversion (SSA)
		"G602": true, // slice access out of bounds (SSA)
		"G407": true, // hardcoded IV (SSA/taint)
	}
	for _, a := range gosecAll {
		if !brokenSSA[a.Name] {
			mychecks = append(mychecks, a)
		}
	}

	checks := make(map[string]bool)
	for _, v := range ignoreChecks {
		checks[v] = true
	}

	mychecks = append(mychecks, filterStaticcheck(checks)...)

	// Стандартные анализаторы passes (printf, shadow, structtag, nilness).
	// Публичные анализаторы (errcheck, bodyclose, gosec).
	// Мой анализатор (osexitcheck).
	// Анализаторы библиотеки staticcheck.
	multichecker.Main(
		mychecks...,
	)
}

func filterStaticcheck(ignoreChecks map[string]bool) []*analysis.Analyzer {
	var result []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		if ignoreChecks[v.Analyzer.Name] {
			continue
		}
		result = append(result, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		if ignoreChecks[v.Analyzer.Name] {
			continue
		}
		result = append(result, v.Analyzer)
	}

	for _, v := range stylecheck.Analyzers {
		if ignoreChecks[v.Analyzer.Name] {
			continue
		}
		result = append(result, v.Analyzer)
	}

	return result
}
