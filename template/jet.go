package template

import (
	"io"

	"github.com/CloudyKit/jet"
)

var set *jet.Set

type Loader struct {
}

func (l Loader) Open(name string) (io.ReadCloser, error) {
	return FS.Open("/assets/" + name)
}

func (l Loader) Exists(name string) (string, bool) {
	_, err := FS.Open("/assets/" + name)
	if err != nil {
		return "", false
	}
	return name, true
}

func nopEscape(w io.Writer, b []byte) {
	_, _ = w.Write(b)
}

func getSet() *jet.Set {
	if set != nil {
		return set
	}
	set = jet.NewSetLoader(nopEscape, &Loader{})
	for name, fn := range CustomFunctions {
		set.AddGlobal(name, fn)
	}
	return set
}
