package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename.go> <function_name>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	targetFunc := os.Args[2]

	fset := token.NewFileSet()

	// We use parser.ParseComments to ensure doc comments above the function are preserved.
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filename, err)
		os.Exit(1)
	}

	found := false

	ast.Inspect(node, func(n ast.Node) bool {
		// Return true to keep walking down the tree

		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Name.Name != targetFunc {
			return true
		}

		found = true
		// Output as Go source code
		err := printer.Fprint(os.Stdout, fset, fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error printing function code: %v\n", err)
		}
		fmt.Println("\n") // Add newlines in case of multiple matches (e.g., methods on different structs)

		// Return false to skip inspecting the interior elements of this function
		return false
	})

	if !found {
		fmt.Fprintf(os.Stderr, "Error: Function '%s' not found in %s\n", targetFunc, filename)
		os.Exit(1)
	}
}
