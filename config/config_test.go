package config

import (
	"github.com/spf13/afero"
	"gs/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() {
	fs.SetTestFs(afero.NewMemMapFs())
}

func TestRead_WithValidModule(t *testing.T) {
	setup()

	_ = fs.WriteFile("go.mod", "module github.com/example/project")

	config, err := read()

	assert.Nil(t, err, "should be nil")
	assert.Equal(t, "github.com/example/project", config.Module, "should be equal")
}

func TestRead_WithInvalidModule(t *testing.T) {
	setup()

	_ = fs.WriteFile("go.mod", "module")

	_, err := read()

	assert.NotNil(t, err, "should not be nil")
}

func TestRead_WithoutModule(t *testing.T) {
	setup()

	_, err := read()

	assert.NotNil(t, err, "should not be nil")
}

func TestReadModule_WithValidModule(t *testing.T) {
	setup()

	_ = fs.WriteFile("go.mod", "module github.com/example/project")

	module, err := readModule()

	assert.Nil(t, err, "should be nil")
	assert.Equal(t, "github.com/example/project", module, "should be equal")
}

func TestReadModule_WithInvalidModule(t *testing.T) {
	setup()

	_ = fs.WriteFile("go.mod", "module")

	module, err := readModule()

	assert.NotNil(t, err, "should not be nil")
	assert.Equal(t, "", module, "should be equal")
}

func TestReadModule_WithoutModule(t *testing.T) {
	setup()

	module, err := readModule()

	assert.NotNil(t, err, "should not be nil")
	assert.Equal(t, "", module, "should be equal")
}

func TestReadModule_WithComplexModule(t *testing.T) {
	setup()

	complexModule := `module github.com/example/project

    require (
        github.com/pkg/errors v0.8.1
        golang.org/x/net v0.0.0-20190620200207-3b0461eec859
        golang.org/x/sys v0.0.0-20190626221950-04f50cda93cb
    )`

	_ = fs.WriteFile("go.mod", complexModule)

	module, err := readModule()

	assert.Nil(t, err, "should be nil")
	assert.Equal(t, "github.com/example/project", module, "should be equal")
}
