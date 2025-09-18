package main

import (
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/parser"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	checkParserFiles()
	checkTransformerFiles()
}

func checkParserFiles() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get working dir:", err)
		return
	}

	parserDir := filepath.Join(dir, "testdata", "parser")
	inputDir := filepath.Join(parserDir, "input")
	outputDir := filepath.Join(parserDir, "output")

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Println("failed to read input dir:", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		inputFilePath := filepath.Join(inputDir, entry.Name())
		outputFilePath := filepath.Join(outputDir, entry.Name())
		fileInfo, err := os.Stat(outputFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("missing output for %s\n", entry.Name())
				generateParserOutputFile(inputFilePath, outputFilePath)
				continue
			}
			fmt.Println("error checking file:", err)
			continue
		}

		fmt.Println("file:", fileInfo.Name())
	}
}

func checkTransformerFiles() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get working dir:", err)
		return
	}

	transformerDir := filepath.Join(dir, "testdata", "transformer")
	inputDir := filepath.Join(transformerDir, "input")
	outputDir := filepath.Join(transformerDir, "output")

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Println("failed to read input dir:", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		inputFilePath := filepath.Join(inputDir, entry.Name())
		outputFilePath := filepath.Join(outputDir, entry.Name())
		fileInfo, err := os.Stat(outputFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("missing output for %s\n", entry.Name())
				generateTransformerOutputFile(inputFilePath, outputFilePath)
				continue
			}
			fmt.Println("error checking file:", err)
			continue
		}

		fmt.Println("file:", fileInfo.Name())
	}
}

func generateParserOutputFile(inputFilePath, outputFilePath string) {
	data, err := os.ReadFile(inputFilePath)
	if err != nil {
		fmt.Println("failed to read input file:", err)
		return
	}

	sourceFile := parser.ParseSourceFile(string(data))

	output, err := json.Marshal(sourceFile)
	if err != nil {
		fmt.Println("failed to marshal AST to JSON:", err)
		return
	}

	if err := os.WriteFile(outputFilePath, output, 0o644); err != nil {
		fmt.Println("failed to write output file:", err)
		return
	}

	fmt.Printf("✨ generated %s\n", outputFilePath)
}

func generateTransformerOutputFile(inputFilePath, outputFilePath string) {
	program := compiler.NewProgram()

	program.ParseSourceFiles([]string{inputFilePath})
	program.TransformSourceFiles()

	output, err := json.Marshal(program.SourceFiles()[0])
	if err != nil {
		fmt.Println("failed to marshal AST to JSON:", err)
		return
	}

	if err := os.WriteFile(outputFilePath, output, 0o644); err != nil {
		fmt.Println("failed to write output file:", err)
		return
	}

	fmt.Printf("generated %s\n", outputFilePath)
}
