package ast

type Diagnostic struct {
	SourceFile *SourceFile
	Location   Location
	Message    string
}

func NewDiagnostic(sourceFile *SourceFile, location Location, message string) *Diagnostic {
	return &Diagnostic{
		SourceFile: sourceFile,
		Location:   location,
		Message:    message,
	}
}
