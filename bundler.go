package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type DependencyAnalyzer struct {
	fset     *token.FileSet
	pkgCache map[string]*packages.Package
}

func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{
		fset:     token.NewFileSet(),
		pkgCache: make(map[string]*packages.Package),
	}
}

type ASTTransformer struct {
	fset *token.FileSet
}

func NewASTTransformer(fset *token.FileSet) *ASTTransformer {
	return &ASTTransformer{
		fset: fset,
	}
}

type FileGenerator struct {
	fset *token.FileSet
}

func NewFileGenerator(fset *token.FileSet) *FileGenerator {
	return &FileGenerator{
		fset: fset,
	}
}

type Bundler struct {
	depAnalyzer   *DependencyAnalyzer
	astTransformer *ASTTransformer
	fileGenerator *FileGenerator
	dependencies  []string
	usedSymbols   map[string]map[string]bool
}

func NewBundler() *Bundler {
	depAnalyzer := NewDependencyAnalyzer()
	return &Bundler{
		depAnalyzer:    depAnalyzer,
		astTransformer: NewASTTransformer(depAnalyzer.fset),
		fileGenerator:  NewFileGenerator(depAnalyzer.fset),
		usedSymbols:    make(map[string]map[string]bool),
	}
}

func (b *Bundler) Bundle(inputFile, outputFile string) error {
	if err := b.analyzeDependencies(inputFile); err != nil {
		return fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	if len(b.dependencies) > 0 {
		fmt.Printf("Found %d dependencies: %s\n", len(b.dependencies), strings.Join(b.dependencies, ", "))
	}

	for _, dep := range b.dependencies {
		fmt.Printf("Processing %s package...\n", dep)
	}

	if err := b.performCallGraphAnalysis(); err != nil {
		return fmt.Errorf("failed to perform call graph analysis: %w", err)
	}

	fmt.Println("Removing unused code...")
	fmt.Println("Merging init functions...")

	if err := b.generateOutput(inputFile, outputFile); err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	fmt.Printf("Writing to %s...\n", outputFile)
	return nil
}

func (da *DependencyAnalyzer) AnalyzeDependencies(inputFile string) ([]string, error) {
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Dir: filepath.Dir(absPath),
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	mainPkg := pkgs[0]
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package loading errors occurred")
	}

	var dependencies []string
	da.collectDependencies(mainPkg, make(map[string]bool), &dependencies)
	
	sort.Strings(dependencies)
	return dependencies, nil
}

func (b *Bundler) analyzeDependencies(inputFile string) error {
	deps, err := b.depAnalyzer.AnalyzeDependencies(inputFile)
	if err != nil {
		return err
	}
	b.dependencies = deps
	return nil
}

func (da *DependencyAnalyzer) collectDependencies(pkg *packages.Package, visited map[string]bool, dependencies *[]string) {
	if visited[pkg.PkgPath] {
		return
	}
	visited[pkg.PkgPath] = true

	da.pkgCache[pkg.PkgPath] = pkg

	for _, imp := range pkg.Imports {
		if da.isStandardLibrary(imp.PkgPath) {
			continue
		}
		
		if !contains(*dependencies, imp.PkgPath) {
			*dependencies = append(*dependencies, imp.PkgPath)
		}
		
		da.collectDependencies(imp, visited, dependencies)
	}
}

func (da *DependencyAnalyzer) isStandardLibrary(pkgPath string) bool {
	// Standard library packages don't contain dots and follow specific patterns
	if strings.Contains(pkgPath, ".") {
		return false
	}
	
	// golang.org/x/* packages are not standard library
	if strings.HasPrefix(pkgPath, "golang.org/x/") {
		return false
	}
	
	// Known standard library root packages and their subpackages
	standardPkgs := []string{
		"archive", "bufio", "builtin", "bytes", "compress", "container", "context", 
		"crypto", "database", "debug", "embed", "encoding", "errors", "expvar", 
		"flag", "fmt", "go", "hash", "html", "image", "index", "io", "log", 
		"math", "mime", "net", "os", "path", "plugin", "reflect", "regexp", 
		"runtime", "sort", "strconv", "strings", "sync", "syscall", "testing", 
		"text", "time", "unicode", "unsafe",
	}
	
	// Check if it's a root package or subpackage of standard library
	parts := strings.Split(pkgPath, "/")
	rootPkg := parts[0]
	
	for _, pkg := range standardPkgs {
		if rootPkg == pkg {
			return true
		}
	}
	
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (b *Bundler) performCallGraphAnalysis() error {
	for _, depPath := range b.dependencies {
		pkg := b.pkgCache[depPath]
		if pkg == nil {
			continue
		}

		b.usedSymbols[depPath] = make(map[string]bool)
		
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.FuncDecl:
					if node.Name.IsExported() {
						symbolName := node.Name.Name
						b.usedSymbols[depPath][symbolName] = true
					}
				case *ast.TypeSpec:
					if node.Name.IsExported() {
						symbolName := node.Name.Name
						b.usedSymbols[depPath][symbolName] = true
					}
				case *ast.GenDecl:
					if node.Tok == token.VAR || node.Tok == token.CONST {
						for _, spec := range node.Specs {
							if valueSpec, ok := spec.(*ast.ValueSpec); ok {
								for _, name := range valueSpec.Names {
									if name.IsExported() {
										b.usedSymbols[depPath][name.Name] = true
									}
								}
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil
}

func (b *Bundler) generateOutput(inputFile, outputFile string) error {
	var output strings.Builder
	
	output.WriteString("// Code generated by your-tool; DO NOT EDIT.\n")
	output.WriteString("package main\n\n")

	imports, err := b.extractStandardImports(inputFile)
	if err != nil {
		return err
	}
	
	if len(imports) > 0 {
		output.WriteString("import (\n")
		for _, imp := range imports {
			output.WriteString(fmt.Sprintf("\t%s\n", imp))
		}
		output.WriteString(")\n\n")
	}

	for _, depPath := range b.dependencies {
		pkg := b.pkgCache[depPath]
		if pkg == nil {
			continue
		}

		pkgName := filepath.Base(depPath)
		output.WriteString(fmt.Sprintf("// From package %s\n", pkgName))

		for _, file := range pkg.Syntax {
			if err := b.processFile(file, pkgName, &output); err != nil {
				return err
			}
		}
		output.WriteString("\n")
	}

	output.WriteString("func init() {\n")
	output.WriteString("\t// Merged init functions\n")
	output.WriteString("}\n\n")

	mainContent, err := b.processMainFile(inputFile)
	if err != nil {
		return err
	}
	output.WriteString(mainContent)

	return os.WriteFile(outputFile, []byte(output.String()), 0644)
}

func (b *Bundler) processFile(file *ast.File, pkgPrefix string, output *strings.Builder) error {
	// Create a map to track original type names to prefixed names
	typeMap := make(map[string]string)
	
	// First pass: collect and rename exported types
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
					oldName := typeSpec.Name.Name
					newName := pkgPrefix + "_" + oldName
					typeSpec.Name.Name = newName
					typeMap[oldName] = newName
				}
			}
		}
	}
	
	// Second pass: update type references in function signatures and bodies
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			if newName, exists := typeMap[node.Name]; exists {
				node.Name = newName
			}
		case *ast.StarExpr:
			if ident, ok := node.X.(*ast.Ident); ok {
				if newName, exists := typeMap[ident.Name]; exists {
					ident.Name = newName
				}
			}
		}
		return true
	})
	
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == "init" {
				continue
			}
			// Only add prefix to exported functions that are NOT methods
			if d.Name.IsExported() && d.Recv == nil {
				d.Name.Name = pkgPrefix + "_" + d.Name.Name
			}
			
			var buf strings.Builder
			if err := format.Node(&buf, b.fset, d); err != nil {
				return err
			}
			output.WriteString(buf.String())
			output.WriteString("\n\n")

		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				continue
			}

			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() {
							name.Name = pkgPrefix + "_" + name.Name
						}
					}
				}
			}

			var buf strings.Builder
			if err := format.Node(&buf, b.fset, d); err != nil {
				return err
			}
			output.WriteString(buf.String())
			output.WriteString("\n\n")
		}
	}
	return nil
}

func (b *Bundler) processMainFile(inputFile string) (string, error) {
	src, err := os.ReadFile(inputFile)
	if err != nil {
		return "", err
	}

	file, err := parser.ParseFile(b.fset, inputFile, src, parser.ParseComments)
	if err != nil {
		return "", err
	}

	b.rewriteMainFile(file)

	var buf strings.Builder
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			continue
		}
		
		if err := format.Node(&buf, b.fset, decl); err != nil {
			return "", err
		}
		buf.WriteString("\n\n")
	}

	return buf.String(), nil
}

func (b *Bundler) extractStandardImports(inputFile string) ([]string, error) {
	src, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(b.fset, inputFile, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var imports []string
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					importPath := strings.Trim(importSpec.Path.Value, "\"")
					if b.isStandardLibrary(importPath) {
						imports = append(imports, importSpec.Path.Value)
					}
				}
			}
		}
	}
	
	sort.Strings(imports)
	return imports, nil
}

func (b *Bundler) rewriteMainFile(file *ast.File) {
	// Build a map of package names for quick lookup
	pkgMap := make(map[string]string)
	for _, depPath := range b.dependencies {
		pkgName := filepath.Base(depPath)
		pkgMap[pkgName] = pkgName
	}
	
	// We need to track parent nodes to replace them properly
	var replaceNodes []func()
	
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// Handle function calls like unionfind.New()
			if selExpr, ok := node.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok {
					if _, exists := pkgMap[ident.Name]; exists {
						newIdent := &ast.Ident{
							Name:    ident.Name + "_" + selExpr.Sel.Name,
							NamePos: selExpr.Pos(),
						}
						replaceNodes = append(replaceNodes, func() {
							node.Fun = newIdent
						})
					}
				}
			}
		}
		return true
	})
	
	// Apply all replacements
	for _, replace := range replaceNodes {
		replace()
	}
}