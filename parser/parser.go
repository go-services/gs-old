package parser

import (
	"bufio"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-services/source"
	"gs/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AnnotatedFile represents a Go file that contains annotations.
type AnnotatedFile struct {
	Path string
	Src  source.Source
}

// commonIgnoreFiles is a list of file names that are commonly used to specify ignored files.
var commonIgnoreFiles = []string{".gitignore", ".ignore", ".gsignore"}

// alwaysIgnorePatterns is a list of patterns that should always be ignored.
// TODO: add the generated files from the generator here.
var alwaysIgnorePatterns = []string{"**/.git"}

// readIgnoredPatterns reads the ignored patterns from a file.
func readIgnoredPatterns(ignoreFilePath string) ([]string, error) {
	exists, err := fs.Exists(ignoreFilePath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	fileContent, err := fs.ReadFile(ignoreFilePath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	var ignoreList []string
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			ignoreList = append(ignoreList, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreList, nil
}

// findIgnoredPatters finds the ignored patterns in a directory.
func findIgnoredPatters(path string, arr1 []os.FileInfo, arr2 []string) []string {
	m := make(map[string]bool)
	for _, item := range arr1 {
		m[item.Name()] = true
	}
	var result []string
	for _, item := range arr2 {
		if _, ok := m[item]; ok {
			ignored, _ := readIgnoredPatterns(filepath.Join(path, item))
			result = append(result, ignored...)
		}
	}

	return result
}

// findGoFiles finds all Go files in a directory and its subdirectories.
func findGoFiles(root string) ([]string, error) {
	var goFiles []string
	ignoreSet := make(map[string]bool)
	var walkDir func(string) error
	var globalIgnoredPatters []string = append(alwaysIgnorePatterns, commonIgnoreFiles...)
	times := 0
	walkDir = func(path string) error {
		times++
		entries, err := fs.ReadDir(path)
		if err != nil {
			return err
		}
		ignoredPatterns := findIgnoredPatters(path, entries, commonIgnoreFiles)
		for _, file := range ignoredPatterns {
			if strings.Contains(file, "**") {
				globalIgnoredPatters = append(globalIgnoredPatters, file)
			}
		}
		for _, entry := range entries {
			fullPath := filepath.Join(path, entry.Name())
			isFileIgnored := false
			for _, ignoredFile := range globalIgnoredPatters {
				if m, _ := doublestar.PathMatch(ignoredFile, entry.Name()); m {
					isFileIgnored = true
				}
			}
			if isFileIgnored {
				continue
			}
			for _, ignoredFile := range ignoredPatterns {
				if m, _ := doublestar.PathMatch(ignoredFile, entry.Name()); m {
					isFileIgnored = true
				}
			}
			if isFileIgnored {
				continue
			}
			if entry.IsDir() {
				err := walkDir(fullPath)
				if err != nil {
					return err
				}
			} else if strings.HasSuffix(entry.Name(), ".go") && !ignoreSet[fullPath] {
				goFiles = append(goFiles, fullPath)
			}
		}
		return nil
	}

	err := walkDir(root)
	if err != nil {
		return nil, err
	}
	return goFiles, nil
}

// containsAnnotation checks if a string contains an annotation.
func containsAnnotation(s string) bool {
	// Define the regular expression
	re := regexp.MustCompile(`@[a-zA-Z0-9]+\([^)]*\)`)

	// Use the MatchString method to check if the string contains the annotation
	return re.MatchString(s)
}

// findGoFilesContainingAnnotations finds all Go files in a directory and its subdirectories that contain annotations.
func findGoFilesContainingAnnotations(path string) ([]AnnotatedFile, error) {
	filePaths, err := findGoFiles(path)
	if err != nil {
		return nil, err
	}
	var sources []AnnotatedFile
	for _, filePath := range filePaths {
		data, err := fs.ReadFile(filePath)
		if !containsAnnotation(data) {
			continue
		}
		if err != nil {
			return nil, err
		}
		src, err := source.New(data)
		if err != nil {
			return nil, err
		}
		sources = append(sources, AnnotatedFile{
			Path: filePath,
			Src:  *src,
		})
	}
	return sources, nil
}

// ParseFiles parses all Go files in a directory and its subdirectories and returns a list of AnnotatedFile.
func ParseFiles(path string) ([]AnnotatedFile, error) {
	return findGoFilesContainingAnnotations(path)
}
