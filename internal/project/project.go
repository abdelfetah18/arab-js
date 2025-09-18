package project

import (
	"arab_js/internal/compiler"
)

type Project struct {
	Program *compiler.Program
}

func NewProject() *Project {
	return &Project{
		Program: compiler.NewProgram(),
	}
}

func (p *Project) UpdateProgram(filePath string, content string) {
	p.Program.UpdateSourceFile(filePath, content)
}
