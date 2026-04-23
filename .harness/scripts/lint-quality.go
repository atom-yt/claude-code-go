// lint-quality checks code quality rules for Go code.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type QualityIssue struct {
	File    string
	Line    int
	Rule    string
	Message string
	Severity string
}

// Rules configuration
var rules = struct {
	MaxFileSize       int  // lines
	NoPrintf         bool
	NoGlobalVars     bool
	MaxFuncLength    int  // lines
	NoExportedErrors bool
}{
	MaxFileSize:       500,
	NoPrintf:         true,
	NoGlobalVars:     true,
	MaxFuncLength:    100,
	NoExportedErrors:  false, // Allow exported errors for this project
}

type fileChecker struct {
	issues []QualityIssue
	fset   *token.FileSet
}

func newFileChecker() *fileChecker {
	return &fileChecker{
		fset: token.NewFileSet(),
	}
}

func (fc *fileChecker) checkFile(filePath string) error {
	node, err := parser.ParseFile(fc.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Check file size
	fc.checkFileSize(filePath)

	// Walk AST to check rules
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			fc.checkPrintCalls(x, filePath)
		case *ast.GenDecl:
			if x.Tok == token.VAR {
				fc.checkGlobalVars(x, filePath)
			}
		case *ast.FuncDecl:
			fc.checkFuncLength(x, filePath)
		}
		return true
	})

	return nil
}

func (fc *fileChecker) checkFileSize(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	lines := strings.Count(string(content), "\n")
	if lines > rules.MaxFileSize {
		fc.issues = append(fc.issues, QualityIssue{
			File:     filePath,
			Line:     1,
			Rule:     "max-file-size",
			Message:  fmt.Sprintf("file too large: %d lines (max %d)", lines, rules.MaxFileSize),
			Severity: "warning",
		})
	}
}

func (fc *fileChecker) checkPrintCalls(expr *ast.CallExpr, filePath string) {
	if rules.NoPrintf {
		if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				funName := sel.Sel.Name
				if ident.Name == "fmt" && (funName == "Print" || funName == "Printf" || funName == "Println") {
					pos := fc.fset.Position(expr.Pos())
					fc.issues = append(fc.issues, QualityIssue{
						File:     filePath,
						Line:     pos.Line,
						Rule:     "no-printf",
						Message:  fmt.Sprintf("use structured logging instead of fmt.%s", funName),
						Severity: "warning",
					})
				}
			}
		}
	}
}

func (fc *fileChecker) checkGlobalVars(decl *ast.GenDecl, filePath string) {
	if rules.NoGlobalVars && decl.Doc == nil {
		// Skip var declarations with comments (likely intentional)
		pos := fc.fset.Position(decl.Pos())
		fc.issues = append(fc.issues, QualityIssue{
			File:     filePath,
			Line:     pos.Line,
			Rule:     "no-global-vars",
			Message:  "avoid global variables, consider passing as parameters",
			Severity: "warning",
		})
	}
}

func (fc *fileChecker) checkFuncLength(fn *ast.FuncDecl, filePath string) {
	if fn.Body != nil {
		start := fc.fset.Position(fn.Body.Pos())
		end := fc.fset.Position(fn.Body.End())
		length := end.Line - start.Line

		if length > rules.MaxFuncLength {
			fc.issues = append(fc.issues, QualityIssue{
				File:     filePath,
				Line:     start.Line,
				Rule:     "max-func-length",
				Message:  fmt.Sprintf("function %s too long: %d lines (max %d)", fn.Name.Name, length, rules.MaxFuncLength),
				Severity: "info",
			})
		}
	}
}

func findGoFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := filepath.Base(path)
			if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func main() {
	goFiles, err := findGoFiles(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	checker := newFileChecker()

	for _, filePath := range goFiles {
		if err := checker.checkFile(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", filePath, err)
		}
	}

	if len(checker.issues) > 0 {
		// Group by severity
		errors := make([]QualityIssue, 0)
		warnings := make([]QualityIssue, 0)
		infos := make([]QualityIssue, 0)

		for _, issue := range checker.issues {
			switch issue.Severity {
			case "error":
				errors = append(errors, issue)
			case "warning":
				warnings = append(warnings, issue)
			case "info":
				infos = append(infos, issue)
			}
		}

		if len(errors) > 0 {
			fmt.Println("Errors:")
			for _, e := range errors {
				fmt.Printf("  %s:%d [%s] %s\n", e.File, e.Line, e.Rule, e.Message)
			}
		}

		if len(warnings) > 0 {
			fmt.Println("Warnings:")
			for _, w := range warnings {
				fmt.Printf("  %s:%d [%s] %s\n", w.File, w.Line, w.Rule, w.Message)
			}
		}

		if len(infos) > 0 {
			fmt.Println("Info:")
			for _, i := range infos {
				fmt.Printf("  %s:%d [%s] %s\n", i.File, i.Line, i.Rule, i.Message)
			}
		}

		fmt.Printf("\nTotal issues: %d (%d errors, %d warnings, %d info)\n",
			len(checker.issues), len(errors), len(warnings), len(infos))

		os.Exit(1)
	}

	fmt.Println("✓ No quality issues found")
	os.Exit(0)
}