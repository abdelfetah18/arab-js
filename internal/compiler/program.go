package compiler

import (
	"arab_js/internal/binder"
	"arab_js/internal/bundled"
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
	printer "arab_js/internal/compiler/emitter"
	"arab_js/internal/compiler/lexer"
	"arab_js/internal/compiler/parser"
	"arab_js/internal/transformer"
	"fmt"
	"os"
	"path/filepath"
)

type ProgramOptions struct {
	Main string
}

type Program struct {
	ProgramOptions ProgramOptions
	Checker        *checker.Checker
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
	libDom := parser.ParseSourceFile(bundled.ReadLibFile(bundled.LibNameDom))
	libES5 := parser.ParseSourceFile(bundled.ReadLibFile(bundled.LibNameES5))
	libDom.IsDeclarationFile = true
	libES5.IsDeclarationFile = true
	p.sourceFiles = append(p.sourceFiles, libDom, libES5)

	for _, filePath := range sourceFilesPaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		parser := parser.NewParser(lexer.NewLexer(string(data)), ast.SourceFileParseOptions{FileName: filepath.Base(filePath), Path: filePath})
		sourceFile := parser.Parse()
		sourceFile.Path = filePath
		sourceFile.Name = filepath.Base(filePath)

		p.filesByPath[filePath] = sourceFile
		p.sourceFiles = append(p.sourceFiles, sourceFile)

		if len(parser.Diagnostics) > 0 {
			p.Diagnostics = append(p.Diagnostics, parser.Diagnostics...)
		}
	}
	return nil
}

func (p *Program) BindSourceFiles() {
	for _, sourceFile := range p.sourceFiles {
		binder.BindSourceFile(sourceFile)
	}
}

func (p *Program) CheckSourceFiles() *binder.NameResolver {
	p.Checker = checker.NewChecker(p)
	p.Checker.Check()
	p.Diagnostics = p.Checker.Diagnostics
	return p.Checker.NameResolver
}

func (p *Program) TransformSourceFiles() {
	transformer.NewTransformer(p).Transform()
}

func (p *Program) EmitSourceFiles(outputDir string) error {
	for _, sourceFile := range p.SourceFiles() {
		if sourceFile.IsDeclarationFile {
			continue
		}

		outputFileName := fmt.Sprintf("%s.js", sourceFile.Name)
		if p.ProgramOptions.Main == sourceFile.Name {
			outputFileName = "index.js"
		}

		_emitter := printer.NewEmitter()
		_emitter.Emit(sourceFile)
		output := _emitter.Writer.Output

		outputFile := filepath.Join(outputDir, outputFileName)
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	}
	return nil
}

func (p *Program) EmitSourceFile(filePath string) string {
	sourceFile := p.filesByPath[filePath]
	if sourceFile.IsDeclarationFile {
		return ""
	}

	_emitter := printer.NewEmitter()
	_emitter.Emit(sourceFile)
	return _emitter.Writer.Output
}

func (p *Program) GetSourceFile(filePath string) *ast.SourceFile {
	return p.filesByPath[filePath]
}

func (p *Program) UpdateSourceFile(filePath string, content string) {
	_, ok := p.filesByPath[filePath]
	if !ok {
		return
	}

	parser := parser.NewParser(lexer.NewLexer(string(content)), ast.SourceFileParseOptions{FileName: filepath.Base(filePath), Path: filePath})
	sourceFile := parser.Parse()
	sourceFile.Path = filePath
	sourceFile.Name = filepath.Base(filePath)

	p.filesByPath[filePath] = sourceFile
	resultIndex := -1
	for index, s := range p.sourceFiles {
		if s.Path == sourceFile.Path {
			resultIndex = index
			break
		}
	}

	if resultIndex >= 0 {
		p.sourceFiles[resultIndex] = sourceFile
	}

	if len(parser.Diagnostics) > 0 {
		p.Diagnostics = parser.Diagnostics
	}
}
