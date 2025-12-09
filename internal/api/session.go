package api

import (
	"arab_js/internal/compiler/ast"
	"arab_js/internal/project"
	"io/fs"
	"os"
	"path/filepath"
)

type Session struct {
	Id              string           `json:"session_id"`
	Project         *project.Project `json:"-"`
	isProjectLoaded bool             `json:"-"`
}

func NewSession(id string) Session {
	return Session{
		Id:              id,
		Project:         project.NewProject(),
		isProjectLoaded: false,
	}
}

func (s *Session) EmitFile(filePath string) string {
	if !s.isProjectLoaded {
		s.loadProject(filePath)
	}

	return s.Project.Program.EmitSourceFile(filePath)
}

func (s *Session) loadProject(filePath string) {
	program := s.Project.Program

	projectPath, _ := findProjectPath(filePath)
	projectFiles, _ := listFilesWithExt(projectPath, ".arts")

	program.Diagnostics = []*ast.Diagnostic{}
	program.ParseSourceFiles(projectFiles)
	// TODO: report diasnostics

	program.Diagnostics = []*ast.Diagnostic{}
	program.CheckSourceFiles()
	// TODO: report diasnostics

	s.isProjectLoaded = true
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
