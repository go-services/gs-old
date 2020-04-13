package watch

import (
	"fmt"
	"gs/service"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

const binaryName = "gs-watcher"

// Builder composes of both runner and watcher. Whenever watcher gets notified, builder starts a build process, and forces the runner to restart
type Builder struct {
	runner  *Runner
	watcher *Watcher
}

// NewBuilder constructs the Builder instance
func NewBuilder(w *Watcher, r *Runner) *Builder {
	return &Builder{watcher: w, runner: r}
}

// Build listens watch events from Watcher and sends messages to Runner
// when new changes are built.
func (b *Builder) Build() {
	go b.registerSignalHandler()
	go func() {
		// used for triggering the first build
		for _, svc := range b.watcher.gsConfig.Services {
			b.watcher.update <- svc
		}
	}()

	for svc := range b.watcher.Wait() {
		err := service.Generate(svc, b.watcher.gsConfig.Module)
		if err != nil {
			log.Println(err)
			continue
		}
		pkg := path.Join(b.watcher.gsConfig.Module, svc.Name, "cmd")
		fileName := generateBinaryName(path.Join(svc.Name, "cmd"))

		log.WithField("service", svc.Name).Info("Building service")

		// build package
		cmd, err := runCommand("go", "build", "-i", "-o", fileName, pkg)
		if err != nil {
			log.Fatalf("Could not run 'go build' command: %s", err)
		}

		if err := cmd.Wait(); err != nil {
			if err := interpretError(err); err != nil {
				log.Println(fmt.Sprintf("An error occurred while building: %s", err))
			} else {
				log.Println("A build error occurred. Please update your code...", err)
			}

			continue
		}
		log.WithField("service", svc.Name).Info("Running service")
		// and start the new process
		b.runner.restart(fileName)
	}
}

func (b *Builder) registerSignalHandler() {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signals
	b.watcher.Close()
	b.runner.Close()
}

// interpretError checks the error, and returns nil if it is
// an exit code 2 error. Otherwise error is returned as it is.
// when a compilation error occurres, it returns with code 2.
func interpretError(err error) error {
	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return err
	}

	status, ok := exiterr.Sys().(syscall.WaitStatus)
	if !ok {
		return err
	}

	if status.ExitStatus() == 2 {
		return nil
	}

	return err
}

func generateBinaryPrefix() string {
	path := os.Getenv("GOPATH")
	if path != "" {
		return fmt.Sprintf("%s/bin/%s", path, binaryName)
	}

	return path
}

func generateBinaryName(packagePath string) string {
	rand.Seed(time.Now().UnixNano())
	packageName := strings.Replace(packagePath, "/", "-", -1)

	return fmt.Sprintf("%s-%s", generateBinaryPrefix(), packageName)
}
