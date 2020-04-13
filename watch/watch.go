package watch

import (
	"gs/config"
	"gs/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/radovskyb/watcher"
)

var log = logrus.WithFields(logrus.Fields{
	"package": "watch",
})

type Watcher struct {
	gsConfig *config.GSConfig
	watcher  *watcher.Watcher
	// when a file gets changed a message is sent to the update channel
	update chan config.ServiceConfig
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
	for _, svc := range w.gsConfig.Services {
		if strings.HasPrefix(pth, getPath(svc.Name)) {
			w.update <- svc
		}
	}
	if pth == viper.GetString(config.GSConfigFileName) {
		w.gsConfig, err = config.ReRead()
		if err != nil {
			return
		}
		for _, svc := range w.gsConfig.Services {
			w.update <- svc
		}
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

	if err := w.watcher.Add(getPath(viper.GetString(config.GSConfigFileName))); err != nil {
		log.Fatalln(err)
	}

	if err := w.watcher.Ignore(".git"); err != nil {
		log.Fatalln(err)
	}
	for _, service := range w.gsConfig.Services {
		if err := w.watcher.AddRecursive(getPath(service.Name)); err != nil {
			log.Fatalln(err)
		}
		if err := w.watcher.Ignore(getPath(path.Join(service.Name, "gen"))); err != nil {
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
func (w *Watcher) Wait() <-chan config.ServiceConfig {
	return w.update
}

// Close closes the fsnotify watcher channel
func (w *Watcher) Close() {
	close(w.update)
}

func getPath(pth string) string {
	if tp := viper.GetString(fs.DebugKey); tp != "" {
		return path.Join(tp, pth)
	}
	return pth
}
func NewWatcher() *Watcher {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}
	return &Watcher{
		update:   make(chan config.ServiceConfig),
		gsConfig: cfg,
		watcher:  watcher.New(),
	}
}
func Run() {
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
