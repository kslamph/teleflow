package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [-w] <file_or_dir_path...>", os.Args[0])
	}

	writeMode := false
	paths := []string{}

	for _, arg := range os.Args[1:] {
		if arg == "-w" {
			writeMode = true
		} else {
			paths = append(paths, arg)
		}
	}

	if len(paths) == 0 {
		log.Fatal("No file or directory paths provided.")
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			log.Printf("Error accessing path %s: %v. Skipping.", path, err)
			continue
		}

		if info.IsDir() {
			err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("Error accessing %s during walk: %v. Skipping.", filePath, err)
					return nil // Continue walking
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
					processFile(filePath, writeMode)
				}
				return nil
			})
			if err != nil {
				log.Printf("Error walking directory %s: %v", path, err)
			}
		} else if strings.HasSuffix(path, ".go") {
			processFile(path, writeMode)
		} else {
			log.Printf("Skipping non-Go file: %s", path)
		}
	}
}

func processFile(filePath string, writeMode bool) {
	fmt.Printf("Processing: %s\n", filePath)
	fset := token.NewFileSet()

	// Read the source file
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return
	}

	// Parse the file. parser.ParseComments is crucial so comments are part of the AST.
	// However, we will later tell the printer to ignore them.
	// Alternatively, we can parse without comments initially, but it's safer
	// to parse with them and then explicitly remove them from the AST node.
	fileNode, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s: %v", filePath, err)
		return
	}

	// The key step: Remove comments from the AST node.
	// Setting fileNode.Comments to nil removes file-level comment groups,
	// but we also need to remove comments attached to individual AST nodes.
	fileNode.Comments = nil

	// Walk the AST and remove comments attached to individual nodes
	ast.Inspect(fileNode, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Field:
			node.Comment = nil
		case *ast.ValueSpec:
			node.Comment = nil
		case *ast.TypeSpec:
			node.Comment = nil
		case *ast.GenDecl:
			node.Doc = nil
		case *ast.FuncDecl:
			node.Doc = nil
		}
		return true
	})

	// You can also iterate and selectively remove if you have complex needs, e.g., keep //go:build
	// var filteredComments []*ast.CommentGroup
	// for _, cg := range fileNode.Comments {
	//  isDirective := false
	//  for _, c := range cg.List {
	//      if strings.HasPrefix(c.Text, "//go:build") || strings.HasPrefix(c.Text, "//go:generate") {
	//          isDirective = true
	//          break
	//      }
	//  }
	//  if isDirective {
	//      filteredComments = append(filteredComments, cg)
	//  }
	// }
	// fileNode.Comments = filteredComments

	// Configure the printer
	// The printer will reformat the code. If you want to try and preserve
	// original formatting as much as possible (minus comments), it's tricky.
	// By default, it will gofmt the output.

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, fileNode); err != nil {
		log.Printf("Error formatting AST for %s: %v", filePath, err)
		return
	}
	cleanedSource := buf.String()

	if writeMode {
		// Make a backup (optional but recommended)
		backupPath := filePath + ".bak"
		if err := os.WriteFile(backupPath, src, 0644); err != nil {
			log.Printf("Warning: Failed to create backup for %s: %v", filePath, err)
		} else {
			fmt.Printf("  Backup created: %s\n", backupPath)
		}

		// Write the cleaned source back to the original file
		if err := os.WriteFile(filePath, []byte(cleanedSource), 0644); err != nil {
			log.Printf("Error writing cleaned file %s: %v", filePath, err)
		} else {
			fmt.Printf("  Successfully removed comments and wrote to %s\n", filePath)
		}
	} else {
		fmt.Println("--- Cleaned Source ---")
		fmt.Println(cleanedSource)
		fmt.Println("----------------------")
	}
}
