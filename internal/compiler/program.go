package compiler

import (
	"arab_js/internal/binder"
	"arab_js/internal/bundled"
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/parser"
	"arab_js/internal/compiler/printer"
	"arab_js/internal/transformer"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type ProgramOptions struct {
	Main string
}

type Program struct {
	ProgramOptions ProgramOptions
	sourceFiles    []*ast.SourceFile
	filesByPath    map[string]*ast.SourceFile

	Diagnostics []*ast.Diagnostic
}

func NewProgram() *Program {
	return &Program{
		ProgramOptions: ProgramOptions{
			Main: "الرئيسية.كود",
		},
		sourceFiles: []*ast.SourceFile{},
		Diagnostics: []*ast.Diagnostic{},
		filesByPath: map[string]*ast.SourceFile{},
	}
}

func (p *Program) SourceFiles() []*ast.SourceFile { return p.sourceFiles }

func (p *Program) ParseSourceFiles(sourceFilesPaths []string) error {
	libFileContent := bundled.ReadLibFile(bundled.LibNameDom)
	p.sourceFiles = append(p.sourceFiles, parser.ParseSourceFile(libFileContent))

	for _, filePath := range sourceFilesPaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		sourceFile := parser.ParseSourceFile(string(data))
		p.filesByPath[filePath] = sourceFile
		p.sourceFiles = append(p.sourceFiles, sourceFile)
	}
	return nil
}

func (p *Program) BindSourceFiles() {
	for _, sourceFile := range p.sourceFiles {
		binder.BindSourceFile(sourceFile)
	}
}

func (p *Program) CheckSourceFiles() *binder.NameResolver {
	c := checker.NewChecker(p)
	c.Check()
	p.Diagnostics = c.Diagnostics
	return c.NameResolver
}

func (p *Program) TransformSourceFiles() {
	transformer.NewTransformer(p).Transform()
}

func (p *Program) WriteSourceFiles(outputDir string) error {
	for _, sourceFile := range p.SourceFiles() {
		outputFileName := fmt.Sprintf("%s.js", sourceFile.Name)
		if p.ProgramOptions.Main == sourceFile.Name {
			outputFileName = "index.js"
		}

		p := printer.NewPrinter()
		p.Write(sourceFile)
		output := p.Writer.Output

		outputFile := filepath.Join(outputDir, outputFileName)
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	}
	return nil
}

func (p *Program) GetSourceFile(filePath string) *ast.SourceFile {
	log.Printf("p.filesByPath=%v\n", p.filesByPath)
	return p.filesByPath[filePath]
}

func (p *Program) UpdateSourceFile(filePath string, content string) {
	_, ok := p.filesByPath[filePath]
	if !ok {
		return
	}

	sourceFile := parser.ParseSourceFile(content)
	p.filesByPath[filePath] = sourceFile
	binder.BindSourceFile(sourceFile)
}
