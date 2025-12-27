package parser

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParser(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")

	parserDir := filepath.Join(repoRoot, "testdata", "parser")
	inputDir := filepath.Join(parserDir, "input")
	outputDir := filepath.Join(parserDir, "output")

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("failed to read input dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			inputFilePath := filepath.Join(inputDir, entry.Name())
			outputFilePath := filepath.Join(outputDir, entry.Name())

			inputBytes, err := os.ReadFile(inputFilePath)
			if err != nil {
				t.Fatalf("failed to read input file %s: %v", inputFilePath, err)
			}

			outputBytes, err := os.ReadFile(outputFilePath)
			if err != nil {
				t.Fatalf("failed to read output file %s: %v", outputFilePath, err)
			}

			sourceFile := ParseSourceFile(string(inputBytes))
			data, err := json.Marshal(sourceFile)
			if err != nil {
				t.Fatalf("failed to marshal AST: %v", err)
			}

			if !bytes.Equal(outputBytes, data) {
				t.Errorf("AST mismatch for %s\nGot:\n%s\nWant:\n%s", entry.Name(), data, outputBytes)
			}
		})
	}
}

func TestTest262ParserTests(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")

	parserDir := filepath.Join(repoRoot, "testdata", "test262-parser-tests")
	passDir := filepath.Join(parserDir, "pass")

	entries, err := os.ReadDir(passDir)
	if err != nil {
		t.Fatalf("failed to read input dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			inputFilePath := filepath.Join(passDir, entry.Name())

			inputBytes, err := os.ReadFile(inputFilePath)
			if err != nil {
				t.Fatalf("failed to read input file %s: %v", inputFilePath, err)
			}

			sourceFile := ParseSourceFile(string(inputBytes))
			_, err = json.Marshal(sourceFile)
			if err != nil {
				t.Fatalf("failed to marshal AST: %v", err)
			}
		})
	}
}
