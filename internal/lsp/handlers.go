package lsp

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"context"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

type Handlers struct {
	FileLoader *compiler.FileLoader
}

func NewHandlers() *Handlers {
	return &Handlers{
		FileLoader: nil,
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
			panic(err)
		}

		logs.Printf("projectFiles=%v\n", projectFiles)

		h.FileLoader = compiler.NewFileLoader(projectFiles)
		h.FileLoader.LoadSourceFiles()
		program := compiler.NewProgram(h.FileLoader.SourceFiles)
		_checker := checker.NewChecker(program)
		_checker.Check()

		if len(_checker.Errors) > 0 {
			for _, message := range _checker.Errors {
				println(message)
			}
			panic("type checking errors")
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
