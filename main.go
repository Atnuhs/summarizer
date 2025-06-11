package main

import (
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/tools/go/packages"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})))
}

func main() {
	println("Hello world")
}

type Bundler struct {
	cfg *packages.Config
}

func NewBuilder() *Bundler {
	return &Bundler{
		cfg: &packages.Config{
			Mode: packages.NeedName |
				packages.NeedFiles |
				packages.NeedSyntax |
				packages.NeedTypes |
				packages.NeedTypesInfo |
				packages.NeedDeps |
				packages.NeedImports,
		},
	}
}

func (b *Bundler) Bundle(dir string) (string, error) {
	b.cfg.Dir = dir
	pkgs, err := packages.Load(b.cfg, ".")
	if err != nil {
		return "", err
	}

	var mainPkg *packages.Package
	internalPkgs := make(map[string]*packages.Package)

	fmt.Println(len(pkgs))
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			mainPkg = pkg
		} else {
			internalPkgs[pkg.PkgPath] = pkg
		}
	}

	return b.merge(mainPkg, internalPkgs)
}

func (b *Bundler) merge(pkg *packages.Package, internalPkgs map[string]*packages.Package) (string, error) {
	slog.Info("merge started", "package", pkg.Name)
	defer slog.Info("merge finished", "package", pkg.Name)

	fmt.Println(pkg.Imports)

	return "hoge", nil
}
