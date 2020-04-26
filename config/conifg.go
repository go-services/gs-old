package config

import (
	"bytes"
	"errors"
	"gs/fs"
	"strconv"

	"github.com/pelletier/go-toml"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithFields(logrus.Fields{"package": "config"})

type AddressConfig struct {
	Url  string `toml:"url"`
	Port int    `toml:"port"`
}

type ServiceConfig struct {
	Http  AddressConfig `toml:"http"`
	Grpc  AddressConfig `toml:"grpc"`
	Debug AddressConfig `toml:"debug"`
}

type GSConfig struct {
	Module          string                   `toml:"-"`
	WatchExtensions []string                 `toml:"watch_extensions"`
	Services        map[string]ServiceConfig `toml:"services"`
}

func Read() (*GSConfig, error) {
	log.Debugf("Reading config...")
	gs, err := fs.ReadFile("gs.toml")
	if err != nil {
		return nil, err
	}
	cfg := &GSConfig{
		Services: make(map[string]ServiceConfig),
	}
	err = toml.Unmarshal([]byte(gs), cfg)
	if err != nil {
		return nil, err
	}
	cfg.Module, err = ReadModule()
	return cfg, err
}

func Write(config GSConfig) error {
	encoded, err := encode(config)
	if err != nil {
		return err
	}
	return fs.WriteFile("gs.toml", encoded)
}
func encode(config GSConfig) (string, error) {
	data := bytes.NewBufferString("")
	encoder := toml.NewEncoder(data)
	err := encoder.Encode(config)
	return data.String(), err
}

func ReadModule() (string, error) {
	mod, err := fs.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	module := modulePath([]byte(mod))
	if module == "" {
		return "", errors.New("could not read the module name")
	}
	return module, nil
}

// Copied from https://github.com/golang/mod/blob/master/modfile/read.go#L882
var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

// ModulePath returns the module path from the gomod file text.
// If it cannot find a module path, it returns an empty string.
// It is tolerant of unrelated problems in the go.mod file.
func modulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}
