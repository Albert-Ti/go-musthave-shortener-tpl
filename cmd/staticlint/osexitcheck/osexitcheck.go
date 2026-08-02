package osexitcheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if s, ok := call.Fun.(*ast.SelectorExpr); ok {
					pkg := pass.TypesInfo.Uses[s.Sel]

					if fn, ok := pkg.(*types.Func); ok && fn.Pkg() != nil {
						if fn.Pkg().Path() == "os" && fn.Name() == "Exit" {
							pass.Reportf(call.Pos(), "прямой вызов os.Exit запрещён в функции main пакета main")
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
