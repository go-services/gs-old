package fs

import (
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/spf13/afero"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetLevel(logrus.DebugLevel)
}

func setup() {
	SetTestFs(afero.NewMemMapFs())
}

func TestAppFs_CreateFolder(t *testing.T) {
	setup()

	err := CreateFolder("abc")
	assert.Nil(t, err, "should be nil")
	b, _ := afero.Exists(testFs, "abc")
	assert.True(t, b, "should be true")
}

func TestAppFs_CreateNestedFolders(t *testing.T) {
	setup()

	err := CreateFolder("abc/123/xyz")
	assert.Nil(t, err, "should be nil")
	b, _ := afero.Exists(testFs, "abc/123/xyz")
	assert.True(t, b, "should be true")
}
func TestAppFs_DeleteFolder(t *testing.T) {
	setup()

	_ = CreateFolder("abc")
	err := DeleteFolder("abc")
	assert.Nil(t, err, "should be nil")
	b, _ := afero.Exists(testFs, "abc")
	assert.False(t, b, "should be false")
}

func TestAppFs_WriteFile(t *testing.T) {
	setup()

	err := WriteFile("abc/123.txt", "Hello, World!")
	assert.Nil(t, err, "should be nil")
	b, _ := afero.Exists(testFs, "abc/123.txt")
	assert.True(t, b, "should be true")
}

func TestAppFs_ReadFile(t *testing.T) {
	setup()

	_ = WriteFile("abc/123.txt", "Hello, World!")
	data, err := ReadFile("abc/123.txt")
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, "Hello, World!", data, "should be equal")
}

func TestAppFs_Exists(t *testing.T) {
	setup()

	_ = CreateFolder("abc")
	b, err := Exists("abc")
	assert.Nil(t, err, "should be nil")
	assert.True(t, b, "should be true")
}

func TestAppFs_Exists_NotExist(t *testing.T) {
	setup()

	b, err := Exists("xyz")
	assert.Nil(t, err, "should be nil")
	assert.False(t, b, "should be false")
}

func TestFs_Walk(t *testing.T) {
	setup()

	// Create directories and files
	_ = CreateFolder("/dir1")
	_ = WriteFile("/dir1/file1.txt", "Hello, World!")
	_ = CreateFolder("/dir2")
	_ = WriteFile("/dir2/file2.txt", "Hello, World!")

	var files []string
	err := Walk("/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})

	assert.Nil(t, err, "should be nil")

	// Check if all files and directories are visited
	expectedFiles := []string{"/", "/dir1", "/dir1/file1.txt", "/dir2", "/dir2/file2.txt"}
	assert.ElementsMatch(t, expectedFiles, files, "should be equal")
}
