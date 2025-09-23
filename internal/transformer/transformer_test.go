package transformer

import (
	"arab_js/internal/binder"
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/parser"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type ProgramStub struct {
	sourceFiles []*ast.SourceFile
}

func (p ProgramStub) SourceFiles() []*ast.SourceFile { return p.sourceFiles }
func (p *ProgramStub) BindSourceFiles() {
	for _, sourceFile := range p.sourceFiles {
		binder.BindSourceFile(sourceFile)
	}
}
func (p *ProgramStub) CheckSourceFiles() *binder.NameResolver {
	c := checker.NewChecker(p)
	c.Check()
	return c.NameResolver
}

func TestTransformer(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")

	transformerDir := filepath.Join(repoRoot, "testdata", "transformer")
	inputDir := filepath.Join(transformerDir, "input")
	outputDir := filepath.Join(transformerDir, "output")

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

			sourceFile := parser.ParseSourceFile(string(inputBytes))
			sourceFile.Name = filepath.Base(outputFilePath)
			NewTransformer(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}}).Transform()

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
