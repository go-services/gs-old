package parser

import (
	"github.com/spf13/afero"
	"gs/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() {
	fs.SetTestFs(afero.NewMemMapFs())
}

func TestReadIgnoreFile_WithValidFile(t *testing.T) {
	setup()
	ignoreFilePath := "testignore.txt"
	_ = fs.WriteFile(ignoreFilePath, "ignore1\nignore2\n")

	ignoreList, err := readIgnoredPatterns(ignoreFilePath)

	assert.Nil(t, err, "should be nil")
	assert.ElementsMatch(t, []string{"ignore1", "ignore2"}, ignoreList, "should be equal")
}

func TestReadIgnoreFile_WithEmptyFile(t *testing.T) {
	setup()
	ignoreFilePath := "testignore.txt"
	_ = fs.WriteFile(ignoreFilePath, "")

	ignoreList, err := readIgnoredPatterns(ignoreFilePath)

	assert.Nil(t, err, "should be nil")
	assert.Empty(t, ignoreList, "should be empty")
}

func TestReadIgnoreFile_WithNonExistentFile(t *testing.T) {
	setup()
	ignoreFilePath := "nonexistent.txt"

	ignoreList, err := readIgnoredPatterns(ignoreFilePath)

	assert.Nil(t, err, "should be nil")
	assert.Nil(t, ignoreList, "should be nil")
}

func TestFindGoFiles_WithValidDirectory(t *testing.T) {
	setup()
	_ = fs.WriteFile("test.go", "package main")
	_ = fs.WriteFile("test2.go", "package main")

	goFiles, err := findGoFiles(".")

	assert.Nil(t, err, "should be nil")
	assert.ElementsMatch(t, []string{"test.go", "test2.go"}, goFiles, "should be equal")
}

func TestFindGoFiles_WithNoGoFiles(t *testing.T) {
	setup()
	_ = fs.WriteFile("test.txt", "Hello, World!")

	goFiles, err := findGoFiles(".")

	assert.Nil(t, err, "should be nil")
	assert.Empty(t, goFiles, "should be empty")
}

func TestFindGoFiles_WithIgnoredFiles(t *testing.T) {
	setup()
	_ = fs.WriteFile("test.go", "package main")
	_ = fs.WriteFile("test2.go", "package main")
	_ = fs.WriteFile(".gsignore", "test2.go")

	goFiles, err := findGoFiles(".")

	assert.Nil(t, err, "should be nil")
	assert.ElementsMatch(t, []string{"test.go"}, goFiles, "should be equal")
}

func TestFindGoFiles_WithNonExistentDirectory(t *testing.T) {
	setup()
	goFiles, err := findGoFiles("nonexistent")

	assert.NotNil(t, err, "should not be nil")
	assert.Nil(t, goFiles, "should be nil")
}
func TestFindGoFiles_WithGlobalIgnoredFiles(t *testing.T) {
	setup()
	_ = fs.WriteFile("abc/tt/test.go", "package main")
	_ = fs.WriteFile("tt/test2.go", "package main")
	_ = fs.WriteFile(".gsignore", "**/tt")

	goFiles, err := findGoFiles(".")

	assert.Nil(t, err, "should be nil")
	assert.ElementsMatch(t, []string{}, goFiles, "should be equal")
}

func TestFindGoFilesContainingAnnotations(t *testing.T) {
	setup()
	_ = fs.WriteFile("abc/tt/test.go", "package main\n// @annotation()")
	_ = fs.WriteFile("abc/tt/test2.go", "package main\n// @annotation(abc='23')")
	_ = fs.WriteFile("tt/test2.go", "package main")

	goFiles, err := findGoFilesContainingAnnotations(".")

	assert.Nil(t, err, "should be nil")
	assert.Len(t, goFiles, 2, "should be equal")
	assert.Equal(t, goFiles[0].Path, "abc/tt/test.go", "should be equal")
	assert.Equal(t, goFiles[0].Src.Package(), "main", "should be equal")
}
