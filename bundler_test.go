package main

import (
	"path/filepath"
	"testing"
)

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
			dir := filepath.Join("testdata/src", tt.testdir)
			absDir, _ := filepath.Abs(dir)

			b := NewBuilder()
			result, err := b.Bundle(absDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("Bundle() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 結果の検証...
			t.Log(result)
		})
	}
}
