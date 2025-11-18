package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

var Level = new(slog.LevelVar)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: Level,
	})))
}

func main() {
	dir := flag.String("dir", ".", "target package directory")
	flag.Parse()

	pkgs, err := loadPackages(*dir)
	if err != nil {
		log.Fatalf("load packages: %v", err)
	}

	// execute summarize
	var raw bytes.Buffer
	if err := Bundle(pkgs, &raw); err != nil {
		log.Fatalf("bundle: %v", err)
	}

	// format with goimports
	formatted, err := imports.Process("main.go", raw.Bytes(), &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  4,
	})
	if err != nil {
		log.Fatalf("goimports: %v", err)
	}
	if _, err := os.Stdout.Write(formatted); err != nil {
		log.Fatalf("write stdout: %v", err)
	}
}

func loadPackages(dir string) ([]*packages.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get abs path of %s", dir)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedDeps |
			packages.NeedModule |
			packages.NeedCompiledGoFiles |
			packages.NeedImports,
		Dir:   absDir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	return pkgs, nil
}
