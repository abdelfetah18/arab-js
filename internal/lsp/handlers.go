package lsp

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/fileloader"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type Handlers struct {
	FileLoader *fileloader.FileLoader
	Server     *lsp.Server
}

func NewHandlers(server *lsp.Server) *Handlers {
	return &Handlers{
		FileLoader: nil,
		Server:     server,
	}
}

func (h *Handlers) OnCompletionHandler(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
	logs.Println("completion: ", req)
	if h.FileLoader == nil {
		return &[]defines.CompletionItem{}, errors.New("no file is open")
	}

	filePath := getPath(req.TextDocument.Uri)
	fileName := filepath.Base(filePath)
	logs.Printf("fileName=%s\n", fileName)
	sourceFile := h.FileLoader.GetSourceFile(fileName)

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

	logs.Printf("findNode=%s\n", findNode.Type)

	// 2. Fetch Symbols From That Location
	keys := getAllEntires(findNode)

	logs.Printf("keys=%v\n", keys)

	return &keys, nil
}

func (h *Handlers) OnDidOpenTextDocumentHandler(ctx context.Context, req *defines.DidOpenTextDocumentParams) (err error) {
	if h.FileLoader == nil {
		projectPath, result := findProjectPath(getPath(req.TextDocument.Uri))
		logs.Printf("projectPath=%s\n", projectPath)
		if !result {
			return errors.New("the current file does not belong to any project")
		}

		projectFiles, err := listFilesWithExt(projectPath, ".كود")
		if err != nil {
			logs.Println(err)
			return err
		}

		logs.Printf("projectFiles=%v\n", projectFiles)

		h.FileLoader = fileloader.NewFileLoader(projectFiles)
		h.FileLoader.LoadSourceFiles()
		program := compiler.NewProgram(h.FileLoader.SourceFiles)
		_checker := checker.NewChecker(program)
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			type Diagnostic struct {
				Range   defines.Range `json:"range"`
				Message string        `json:"message"`
			}

			type PublishDiagnosticsParams struct {
				Uri         defines.DocumentUri `json:"uri"`
				Version     int                 `json:"version"`
				Diagnostics []Diagnostic        `json:"diagnostics"`
			}

			diagnostics := make([]Diagnostic, 0, len(_checker.Diagnostics))
			for _, diagnostic := range _checker.Diagnostics {
				start, err := indexToPosition(getPath(req.TextDocument.Uri), diagnostic.Location.Pos)
				if err != nil {
					start = defines.Position{Line: 0, Character: 0}
				}

				end, err := indexToPosition(getPath(req.TextDocument.Uri), diagnostic.Location.End)
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
				println(diagnostic.Message)
			}

			params := PublishDiagnosticsParams{
				Uri:         req.TextDocument.Uri,
				Version:     req.TextDocument.Version,
				Diagnostics: diagnostics,
			}

			payload, err := json.Marshal(params)
			if err != nil {
				logs.Println(err) // or handle gracefully
				return err
			}

			h.Server.SendNotification("textDocument/publishDiagnostics", payload)
			logs.Println("type checking errors")
		}
	}

	return nil
}

func getPath(uri defines.DocumentUri) string {
	enEscapeUrl, _ := url.QueryUnescape(string(uri))
	return enEscapeUrl[6:]
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

	var lineNum, runeCount uint
	var offset uint

	for i := 0; i < len(data); {
		if lineNum == pos.Line {
			// Count runes until reaching the desired character
			for j := i; j < len(data); {
				if runeCount == pos.Character {
					return uint(j), nil
				}
				r, size := utf8.DecodeRune(data[j:])
				if r == '\n' || r == utf8.RuneError && size == 1 {
					break // end of line or invalid rune
				}
				runeCount++
				j += size
			}
			// Clamp to end of line
			return uint(len(data[:i]) + len(string(data[i:strings.IndexByte(string(data[i:]), '\n')]))), nil
		}

		// Consume a line
		r, size := utf8.DecodeRune(data[i:])
		i += size
		offset += uint(size)
		if r == '\n' {
			lineNum++
			runeCount = 0
		}
	}

	return 0, errors.New("position out of range")
}

func getAllEntires(node *ast.Node) []defines.CompletionItem {
	var currentScope *ast.Scope = getPrentContainer(node)
	keys := []defines.CompletionItem{}

	for currentScope != nil {
		d := defines.CompletionItemKindText
		for k := range currentScope.Locals {
			keys = append(keys, defines.CompletionItem{
				Label:      "code",
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
