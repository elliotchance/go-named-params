package main

import (
	"go/ast"
	"go/token"
	"./parser"
	"fmt"
)

func main() {
	// src is the input for which we want to print the AST.
	src := `
	package main

	func anon(a int, b int) {
	}

	func named(a: int, b: int) {
	}

func main() {
	named(a: 3 + 2, b: 5)
}
`

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}

	// Print the AST.
	ast.Print(fset, f)

	fmt.Printf("%s\n", parser.RenderFile(f, fset))
}
