package parser

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gs/fs"
	"os"
	"testing"
)

func TestFindApis(t *testing.T) {
	fs.SetTestFs(afero.NewOsFs())
	os.Chdir("../testdata/abc")
	files, _ := ParseFiles(".")
	apis, _ := FindServices(files)
	assert.Equal(t, 1, len(apis), "should be equal")
}
