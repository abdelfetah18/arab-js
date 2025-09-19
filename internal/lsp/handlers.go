package lsp

import (
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/project"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type Handlers struct {
	Project *project.Project
	Server  *lsp.Server

	version int
}

func NewHandlers(server *lsp.Server) *Handlers {
	return &Handlers{
		Project: nil,
		Server:  server,
		version: 0,
	}
}

func (h *Handlers) openProjectIfNotAlreadyOpen(uri defines.DocumentUri) {
	if h.Project != nil {
		return
	}

	h.Project = project.NewProject()
	program := h.Project.Program

	projectPath, _ := findProjectPath(getPath(uri))
	projectFiles, _ := listFilesWithExt(projectPath, ".كود")

	program.Diagnostics = []*ast.Diagnostic{}
	program.ParseSourceFiles(projectFiles)
	if h.reportDiagnosticErrors(program, uri) {
		return
	}

	program.Diagnostics = []*ast.Diagnostic{}
	program.CheckSourceFiles()
	h.reportDiagnosticErrors(program, uri)
}

func (h *Handlers) OnCompletionHandler(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
	h.openProjectIfNotAlreadyOpen(req.TextDocument.Uri)

	filePath := getPath(req.TextDocument.Uri)
	sourceFile := h.Project.Program.GetSourceFile(filePath)

	if sourceFile == nil {
		return nil, errors.New("sourceFile not found")
	}

	position, err := positionToIndex(filePath, req.Position) // Get Position from (Line, Character)
	if err != nil {
		return nil, err
	}

	// 1. Locate Node By Position Line and Character
	var findNode *ast.Node = nil
	var current *ast.Node = sourceFile.AsNode()
	var next *ast.Node = nil

	visitAll := func(n *ast.Node) bool {
		if position >= n.Location.Pos && position <= n.Location.End {
			next = n
			return true
		}

		return false
	}

	for {
		found := current.ForEachChild(visitAll)
		if next == nil && !found {
			findNode = current
			break
		}

		current = next
		next = nil
	}

	// 2. Fetch Symbols From That Location
	keys := getAllEntires(findNode)

	return &keys, nil
}

func (h *Handlers) OnDidOpenTextDocumentHandler(ctx context.Context, req *defines.DidOpenTextDocumentParams) (err error) {
	h.openProjectIfNotAlreadyOpen(req.TextDocument.Uri)
	h.version = req.TextDocument.Version
	return nil
}

func (h *Handlers) OnDidChangeTextDocument(ctx context.Context, req *defines.DidChangeTextDocumentParams) (err error) {
	h.version = req.TextDocument.Version

	h.openProjectIfNotAlreadyOpen(req.TextDocument.Uri)
	if len(req.ContentChanges) == 0 {
		return errors.New("document didnt change")
	}
	str := req.ContentChanges[len(req.ContentChanges)-1].Text.(string)
	program := h.Project.Program
	program.Diagnostics = []*ast.Diagnostic{}
	h.Project.UpdateProgram(getPath(req.TextDocument.Uri), str)
	if h.reportDiagnosticErrors(program, req.TextDocument.Uri) {
		return nil
	}

	program.Diagnostics = []*ast.Diagnostic{}
	program.CheckSourceFiles()
	h.reportDiagnosticErrors(program, req.TextDocument.Uri)
	log.Printf("CheckSourceFiles: reportDiagnosticErrors=%d\n", len(program.Diagnostics))
	return nil
}

func (h *Handlers) reportDiagnosticErrors(program *compiler.Program, uri defines.DocumentUri) bool {
	type Diagnostic struct {
		Range   defines.Range `json:"range"`
		Message string        `json:"message"`
	}

	type PublishDiagnosticsParams struct {
		Uri         defines.DocumentUri `json:"uri"`
		Version     int                 `json:"version"`
		Diagnostics []Diagnostic        `json:"diagnostics"`
	}

	diagnostics := []Diagnostic{}
	for _, diagnostic := range program.Diagnostics {
		start, err := indexToPosition(getPath(uri), diagnostic.Location.Pos)
		if err != nil {
			start = defines.Position{Line: 0, Character: 0}
		}

		end, err := indexToPosition(getPath(uri), diagnostic.Location.End)
		if err != nil {
			end = defines.Position{Line: 0, Character: 0}
		}

		diagnostics = append(diagnostics, Diagnostic{
			Range: defines.Range{
				Start: start,
				End:   end,
			},
			Message: diagnostic.Message,
		})

		log.Printf("message=%s\n", diagnostic.Message)
	}

	params := PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: diagnostics,
		Version:     h.version,
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return false
	}

	h.Server.SendNotification("textDocument/publishDiagnostics", payload)

	return len(program.Diagnostics) > 0
}

func getPath(uri defines.DocumentUri) string {
	enEscapeUrl, _ := url.QueryUnescape(string(uri))
	return filepath.Join(enEscapeUrl[6:])
}

func findProjectPath(startPath string) (string, bool) {
	return forEachAncestorPath(startPath, func(directory string) (string, bool) {
		target := filepath.Join(directory, "رزمة.تعريف")
		if _, err := os.Stat(target); err == nil {
			return directory, true
		}
		return "", false
	})
}

func forEachAncestorPath(
	directory string,
	callback func(directory string) (resultPath string, stop bool),
) (string, bool) {
	dir := filepath.Clean(directory)

	for {
		result, stop := callback(dir)
		if stop {
			return result, true
		}

		parent := filepath.Dir(dir)
		if parent == dir { // reached root
			return "", false
		}
		dir = parent
	}
}

func listFilesWithExt(root, ext string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // stop if there's a problem accessing the path
		}
		if !d.IsDir() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func positionToIndex(filePath string, pos defines.Position) (uint, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	line := 0
	character := 0
	for i := 0; i < len(data); {
		ch, size := utf8.DecodeRune(data[i:])
		i += size
		character += size
		if ch == '\n' {
			line += 1
			character = 0
		}

		if line == int(pos.Line) && character == int(pos.Character) {
			return uint(i), nil
		}
	}

	return 0, errors.New("position out of range")
}

func getAllEntires(node *ast.Node) []defines.CompletionItem {
	var currentScope *ast.Scope = getPrentContainer(node)
	keys := []defines.CompletionItem{}

	for currentScope != nil {
		for k, symbol := range currentScope.Locals {
			label := "code"
			d := defines.CompletionItemKindText
			switch symbol.Node.Type {
			case ast.NodeTypeFunctionDeclaration:
				label = "function"
				d = defines.CompletionItemKindFunction
			case ast.NodeTypeVariableDeclaration:
				d = defines.CompletionItemKindVariable
			}
			keys = append(keys, defines.CompletionItem{
				Label:      label,
				Kind:       &d,
				InsertText: &k,
			})
		}
		currentScope = currentScope.Parent
	}

	return keys
}

func getPrentContainer(node *ast.Node) *ast.Scope {
	if node.Type == ast.NodeTypeBlockStatement && node.Parent.Type == ast.NodeTypeFunctionDeclaration {
		return node.Parent.AsFunctionDeclaration().Scope
	}

	switch node.Type {
	case ast.NodeTypeSourceFile:
		return node.AsSourceFile().Scope
	case ast.NodeTypeBlockStatement:
		return node.AsBlockStatement().Scope
	case ast.NodeTypeFunctionDeclaration:
		return node.AsFunctionDeclaration().Scope
	default:
		return getPrentContainer(node.Parent)
	}
}

func indexToPosition(filePath string, idx uint) (defines.Position, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return defines.Position{}, err
	}
	if idx > uint(len(data)) {
		return defines.Position{}, errors.New("index out of range")
	}

	var lineNum, runeCount, byteCount uint

	for i := 0; i < len(data); {
		if byteCount == idx {
			return defines.Position{Line: lineNum, Character: runeCount}, nil
		}

		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// invalid rune, but still move one byte forward
			size = 1
		}

		if r == '\n' {
			lineNum++
			runeCount = 0
		} else {
			runeCount++
		}

		i += size
		byteCount += uint(size)

		// If idx falls inside the rune boundary (shouldn’t in UTF-8 safe input),
		// treat it as if pointing to that rune.
		if byteCount > idx {
			return defines.Position{Line: lineNum, Character: runeCount}, nil
		}
	}

	// idx == len(data) (EOF case)
	if idx == uint(len(data)) {
		return defines.Position{Line: lineNum, Character: runeCount}, nil
	}

	return defines.Position{}, errors.New("index out of range")
}
