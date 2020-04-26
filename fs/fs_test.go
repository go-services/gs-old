package fs

import (
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
	testFs = afero.NewMemMapFs()
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
