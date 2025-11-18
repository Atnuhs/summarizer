package main

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestMain(m *testing.M) {
	Level.Set(slog.LevelDebug)
	os.Exit(m.Run())
}

func loadTestPackage(t *testing.T, dir string) []*packages.Package {
	t.Helper()
	pkgs, err := loadPackages(filepath.Join("testdata/src", dir))
	if err != nil {
		t.Fatal(err)
	}
	return pkgs
}

func TestBundler(t *testing.T) {
	tests := []struct {
		name    string
		testdir string
		wantErr bool
	}{
		{
			name:    "no dependencies",
			testdir: "no-deps",
		},
		{
			name:    "single dependencies",
			testdir: "single-deps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			pkgs := loadTestPackage(t, tt.testdir)

			// execute
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			err := Bundle(pkgs, buf)

			// validate
			if (err != nil) != tt.wantErr {
				t.Errorf("Bundle() error = %v, wantErr %v", err, tt.wantErr)
			}
			// 結果の検証...
			t.Log(buf.String())
		})
	}
}
