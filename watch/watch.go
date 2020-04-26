package watch

import (
	"gs/config"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/radovskyb/watcher"
)

var log = logrus.WithFields(logrus.Fields{
	"package": "watch",
})

type Watcher struct {
	gsConfig *config.GSConfig
	watcher  *watcher.Watcher
	// when a file gets changed a message is sent to the update channel
	update chan string
}

func (w *Watcher) handleUpdate(event watcher.Event) {
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	pth, err := filepath.Rel(currentPath, event.Path)
	if err != nil {
		log.Fatalln(err)
	}
	mustWatch := false
	ext := filepath.Ext(pth)
	for _, v := range append([]string{".go"}, w.gsConfig.WatchExtensions...) {
		if v == ext {
			mustWatch = true
		}
	}
	if pth == "gs.toml" {
		mustWatch = true
	}
	if !mustWatch {
		return
	}
	for name := range w.gsConfig.Services {
		if strings.HasPrefix(pth, name) {
			w.update <- name
			return
		}
	}

	// something outside of any service changed reload all of them
	w.gsConfig, err = config.Read()
	if err != nil {
		return
	}
	for name := range w.gsConfig.Services {
		w.update <- name
	}
}

func (w *Watcher) watchLoop() {
	for {
		select {
		case event := <-w.watcher.Event:
			if !event.IsDir() {
				w.handleUpdate(event)
			}
		case err := <-w.watcher.Error:
			log.Fatalln(err)
		case <-w.watcher.Closed:
			return
		}
	}
}

func (w *Watcher) Watch() {
	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	// If SetMaxEvents is not set, the default is to send all events.
	w.watcher.SetMaxEvents(10)

	runner := NewRunner()

	go w.watchLoop()

	if err := w.watcher.AddRecursive("."); err != nil {
		log.Fatalln(err)
	}

	if err := w.watcher.Ignore(".git"); err != nil {
		log.Fatalln(err)
	}

	for name := range w.gsConfig.Services {
		if err := w.watcher.Ignore(path.Join(name, "gen")); err != nil {
			log.Fatalln(err)
		}
	}

	go func() {
		time.Sleep(1 * time.Second)
		runner.Run()
	}()
	if err := w.watcher.Start(time.Millisecond * 50); err != nil {
		log.Fatalln(err)
	}
}

// Wait waits for the latest messages
func (w *Watcher) Wait() <-chan string {
	return w.update
}

// Close closes the fsnotify watcher channel
func (w *Watcher) Close() {
	close(w.update)
}

func NewWatcher() *Watcher {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}
	return &Watcher{
		update:   make(chan string),
		gsConfig: cfg,
		watcher:  watcher.New(),
	}
}
func Run(port int) {
	r := NewRunner()
	w := NewWatcher()
	// wait for build and run the binary with given params
	go r.Run()

	b := NewBuilder(w, r)

	// build given package
	go b.Build()

	// listen for further changes
	go w.Watch()

	r.Wait()
}
