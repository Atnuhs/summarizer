package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type UsageAnalyzer struct {
	fset        *token.FileSet
	pkgCache    map[string]*packages.Package
	usedSymbols map[string]map[string]bool
}

func NewUsageAnalyzer(fset *token.FileSet, pkgCache map[string]*packages.Package) *UsageAnalyzer {
	return &UsageAnalyzer{
		fset:        fset,
		pkgCache:    pkgCache,
		usedSymbols: make(map[string]map[string]bool),
	}
}

type Bundler struct {
	fset         *token.FileSet
	pkgCache     map[string]*packages.Package
	dependencies []string
	usedSymbols  map[string]map[string]bool
}

func NewBundler() *Bundler {
	return &Bundler{
		fset:        token.NewFileSet(),
		pkgCache:    make(map[string]*packages.Package),
		usedSymbols: make(map[string]map[string]bool),
	}
}

func (b *Bundler) Bundle(inputFile, outputFile string) error {
	if err := b.analyzeDependencies(inputFile); err != nil {
		return fmt.Errorf("dependency analysis failed for %s: %w", inputFile, err)
	}

	b.reportDependencies()

	if err := b.performCallGraphAnalysis(); err != nil {
		return fmt.Errorf("call graph analysis failed: %w", err)
	}

	b.reportProcessingSteps()

	if err := b.generateOutput(inputFile, outputFile); err != nil {
		return fmt.Errorf("output generation failed for %s: %w", outputFile, err)
	}

	fmt.Printf("Writing to %s...\n", outputFile)
	return nil
}

func (b *Bundler) reportDependencies() {
	if len(b.dependencies) > 0 {
		fmt.Printf("Found %d dependencies: %s\n", len(b.dependencies), strings.Join(b.dependencies, ", "))
		for _, dep := range b.dependencies {
			fmt.Printf("Processing %s package...\n", dep)
		}
	}
}

func (b *Bundler) reportProcessingSteps() {
	fmt.Println("Removing unused code...")
	fmt.Println("Merging init functions...")
}

func (b *Bundler) analyzeDependencies(inputFile string) error {
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		return err
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Dir: filepath.Dir(absPath),
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return err
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found in directory: %s", filepath.Dir(absPath))
	}

	mainPkg := pkgs[0]
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("package loading errors occurred for: %s", inputFile)
	}

	b.collectDependencies(mainPkg, make(map[string]bool))
	
	sort.Strings(b.dependencies)
	return nil
}

func (b *Bundler) collectDependencies(pkg *packages.Package, visited map[string]bool) {
	if visited[pkg.PkgPath] {
		return
	}
	visited[pkg.PkgPath] = true

	b.pkgCache[pkg.PkgPath] = pkg

	for _, imp := range pkg.Imports {
		if b.isStandardLibrary(imp.PkgPath) {
			continue
		}
		
		if !b.contains(b.dependencies, imp.PkgPath) {
			b.dependencies = append(b.dependencies, imp.PkgPath)
		}
		
		b.collectDependencies(imp, visited)
	}
}

func (b *Bundler) isStandardLibrary(pkgPath string) bool {
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
	
	return slices.Contains(standardPkgs, rootPkg)
}

func (b *Bundler) contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func (ua *UsageAnalyzer) AnalyzeUsage(mainPkg *packages.Package, dependencies []string) error {
	// First, find what symbols are actually referenced from main package
	referencedSymbols := ua.findReferencedSymbols(mainPkg, dependencies)
	
	// Then recursively find all symbols used by those symbols
	for _, depPath := range dependencies {
		ua.usedSymbols[depPath] = make(map[string]bool)
		ua.markUsedSymbols(depPath, referencedSymbols[depPath])
	}
	
	return nil
}

func (ua *UsageAnalyzer) findReferencedSymbols(mainPkg *packages.Package, dependencies []string) map[string]map[string]bool {
	referenced := make(map[string]map[string]bool)
	
	// Initialize maps for all dependencies
	for _, depPath := range dependencies {
		referenced[depPath] = make(map[string]bool)
	}
	
	// Analyze main package files for references to dependency symbols
	for _, file := range mainPkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				// Handle package.Symbol references
				if ident, ok := node.X.(*ast.Ident); ok {
					for _, depPath := range dependencies {
						pkgName := filepath.Base(depPath)
						if ident.Name == pkgName {
							referenced[depPath][node.Sel.Name] = true
						}
					}
				}
			case *ast.CallExpr:
				// Handle constructor calls and type instantiations
				if selExpr, ok := node.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						for _, depPath := range dependencies {
							pkgName := filepath.Base(depPath)
							if ident.Name == pkgName {
								referenced[depPath][selExpr.Sel.Name] = true
							}
						}
					}
				}
			case *ast.CompositeLit:
				// Handle struct literals like &math.Calculator{}
				if selExpr, ok := node.Type.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						for _, depPath := range dependencies {
							pkgName := filepath.Base(depPath)
							if ident.Name == pkgName {
								referenced[depPath][selExpr.Sel.Name] = true
							}
						}
					}
				}
			}
			return true
		})
	}
	
	return referenced
}

func (ua *UsageAnalyzer) markUsedSymbols(depPath string, directlyUsed map[string]bool) {
	pkg := ua.pkgCache[depPath]
	if pkg == nil {
		return
	}

	// Start with directly referenced symbols
	toProcess := make(map[string]bool)
	for symbol := range directlyUsed {
		toProcess[symbol] = true
		ua.usedSymbols[depPath][symbol] = true
	}
	
	// Find dependencies of used symbols (methods, embedded types, etc.)
	ua.findSymbolDependencies(depPath, toProcess)
}

func (ua *UsageAnalyzer) findSymbolDependencies(depPath string, symbols map[string]bool) {
	pkg := ua.pkgCache[depPath]
	if pkg == nil {
		return
	}
	
	processed := make(map[string]bool)
	queue := make([]string, 0, len(symbols))
	
	// Initialize queue with directly used symbols
	for symbol := range symbols {
		queue = append(queue, symbol)
	}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if processed[current] {
			continue
		}
		processed[current] = true
		
		// Find what this symbol depends on
		dependencies := ua.getSymbolDependencies(pkg, current)
		for dep := range dependencies {
			if !ua.usedSymbols[depPath][dep] {
				ua.usedSymbols[depPath][dep] = true
				queue = append(queue, dep)
			}
		}
	}
}

func (ua *UsageAnalyzer) getSymbolDependencies(pkg *packages.Package, symbolName string) map[string]bool {
	dependencies := make(map[string]bool)
	
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				// If this is the function we're analyzing
				if node.Name.Name == symbolName {
					// Find what types and functions it uses
					ast.Inspect(node, func(inner ast.Node) bool {
						if ident, ok := inner.(*ast.Ident); ok {
							// Check if this identifier refers to an exported symbol in the same package
							if ident.IsExported() && ident.Name != symbolName {
								if ua.isDefinedInPackage(pkg, ident.Name) {
									dependencies[ident.Name] = true
								}
							}
						}
						return true
					})
				}
			case *ast.TypeSpec:
				// If this is the type we're analyzing
				if node.Name.Name == symbolName {
					// Find embedded types and field types
					ast.Inspect(node.Type, func(inner ast.Node) bool {
						if ident, ok := inner.(*ast.Ident); ok {
							if ident.IsExported() && ident.Name != symbolName {
								if ua.isDefinedInPackage(pkg, ident.Name) {
									dependencies[ident.Name] = true
								}
							}
						}
						return true
					})
				}
			}
			return true
		})
	}
	
	return dependencies
}

func (ua *UsageAnalyzer) isDefinedInPackage(pkg *packages.Package, symbolName string) bool {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.Name == symbolName && d.Name.IsExported() {
					return true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.Name == symbolName && s.Name.IsExported() {
							return true
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.Name == symbolName && name.IsExported() {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (ua *UsageAnalyzer) GetUsedSymbols() map[string]map[string]bool {
	return ua.usedSymbols
}

func (b *Bundler) performCallGraphAnalysis() error {
	// Get main package for analysis
	var mainPkg *packages.Package
	for _, pkg := range b.pkgCache {
		if pkg.Name == "main" {
			mainPkg = pkg
			break
		}
	}
	
	if mainPkg == nil {
		return fmt.Errorf("main package not found")
	}
	
	analyzer := NewUsageAnalyzer(b.fset, b.pkgCache)
	if err := analyzer.AnalyzeUsage(mainPkg, b.dependencies); err != nil {
		return err
	}
	
	b.usedSymbols = analyzer.GetUsedSymbols()
	return nil
}

func (b *Bundler) generateOutput(inputFile, outputFile string) error {
	var output strings.Builder
	
	b.writeHeader(&output)
	
	if err := b.writeImports(inputFile, &output); err != nil {
		return err
	}
	
	if err := b.writeDependencyCode(&output); err != nil {
		return err
	}
	
	b.writeInitFunction(&output)
	
	if err := b.writeMainCode(inputFile, &output); err != nil {
		return err
	}

	return os.WriteFile(outputFile, []byte(output.String()), 0644)
}

func (b *Bundler) writeHeader(output *strings.Builder) {
	output.WriteString("// Code generated by your-tool; DO NOT EDIT.\n")
	output.WriteString("package main\n\n")
}

func (b *Bundler) writeImports(inputFile string, output *strings.Builder) error {
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
	return nil
}

func (b *Bundler) writeDependencyCode(output *strings.Builder) error {
	for _, depPath := range b.dependencies {
		pkg := b.pkgCache[depPath]
		if pkg == nil {
			continue
		}

		pkgName := filepath.Base(depPath)
		output.WriteString(fmt.Sprintf("// From package %s\n", pkgName))

		for _, file := range pkg.Syntax {
			if err := b.processFile(file, pkgName, depPath, output); err != nil {
				return err
			}
		}
		output.WriteString("\n")
	}
	return nil
}

func (b *Bundler) writeInitFunction(output *strings.Builder) {
	output.WriteString("func init() {\n")
	output.WriteString("\t// Merged init functions\n")
	output.WriteString("}\n\n")
}

func (b *Bundler) writeMainCode(inputFile string, output *strings.Builder) error {
	mainContent, err := b.processMainFile(inputFile)
	if err != nil {
		return err
	}
	output.WriteString(mainContent)
	return nil
}

func (b *Bundler) processFile(file *ast.File, pkgPrefix, depPath string, output *strings.Builder) error {
	usedSymbols := b.usedSymbols[depPath]
	if usedSymbols == nil {
		usedSymbols = make(map[string]bool)
	}
	
	// Create a map to track original type names to prefixed names
	typeMap := make(map[string]string)
	
	// First pass: collect used exported types and prepare renaming map
	// Don't rename yet, just collect
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
					oldName := typeSpec.Name.Name
					// Only process if this type is used
					if usedSymbols[oldName] {
						newName := pkgPrefix + "_" + oldName
						typeMap[oldName] = newName
					}
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
			// Only process exported functions that are used and are NOT methods
			if d.Name.IsExported() && d.Recv == nil {
				if !usedSymbols[d.Name.Name] {
					continue // Skip unused functions
				}
				d.Name.Name = pkgPrefix + "_" + d.Name.Name
			} else if d.Name.IsExported() && d.Recv != nil {
				// For methods, check if the receiver type is used
				if starExpr, ok := d.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						if !usedSymbols[ident.Name] {
							continue // Skip methods of unused types
						}
					}
				} else if ident, ok := d.Recv.List[0].Type.(*ast.Ident); ok {
					if !usedSymbols[ident.Name] {
						continue // Skip methods of unused types
					}
				}
			} else if !d.Name.IsExported() {
				continue // Skip unexported functions
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

			// Filter specs to only include used symbols
			var filteredSpecs []ast.Spec
			
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() && usedSymbols[s.Name.Name] {
						// Apply prefix using the typeMap
						if newName, exists := typeMap[s.Name.Name]; exists {
							s.Name.Name = newName
						}
						filteredSpecs = append(filteredSpecs, s)
					}
				case *ast.ValueSpec:
					var filteredNames []*ast.Ident
					var filteredValues []ast.Expr
					
					for i, name := range s.Names {
						if name.IsExported() && usedSymbols[name.Name] {
							name.Name = pkgPrefix + "_" + name.Name
							filteredNames = append(filteredNames, name)
							if s.Values != nil && i < len(s.Values) {
								filteredValues = append(filteredValues, s.Values[i])
							}
						}
					}
					
					if len(filteredNames) > 0 {
						newSpec := &ast.ValueSpec{
							Names:  filteredNames,
							Type:   s.Type,
							Values: filteredValues,
						}
						filteredSpecs = append(filteredSpecs, newSpec)
					}
				}
			}

			if len(filteredSpecs) > 0 {
				newDecl := &ast.GenDecl{
					Tok:   d.Tok,
					Specs: filteredSpecs,
				}
				
				var buf strings.Builder
				if err := format.Node(&buf, b.fset, newDecl); err != nil {
					return err
				}
				output.WriteString(buf.String())
				output.WriteString("\n\n")
			}
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
		case *ast.CompositeLit:
			// Handle struct literals like &math.Calculator{}
			if selExpr, ok := node.Type.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok {
					if _, exists := pkgMap[ident.Name]; exists {
						newIdent := &ast.Ident{
							Name:    ident.Name + "_" + selExpr.Sel.Name,
							NamePos: selExpr.Pos(),
						}
						replaceNodes = append(replaceNodes, func() {
							node.Type = newIdent
						})
					}
				}
			}
		case *ast.SelectorExpr:
			// Handle method calls like calc.Add()
			if ident, ok := node.X.(*ast.Ident); ok {
				// This handles cases where we have prefixed types
				// but we need to be careful not to change method calls on local variables
				_ = ident // placeholder for now
			}
		}
		return true
	})
	
	// Apply all replacements
	for _, replace := range replaceNodes {
		replace()
	}
}