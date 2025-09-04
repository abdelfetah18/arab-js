package ast

func IsExternalModule(sourceFile *SourceFile) bool {
	return sourceFile.ExternalModuleIndicator != nil
}
