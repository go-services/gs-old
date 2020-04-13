package watch

import (
	"gs/fs"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/viper"
)

// Runner listens for the change events and depending on that kills
// the obsolete process, and runs a new one
type Runner struct {
	start chan string
	done  chan struct{}
	cmds  map[string]*exec.Cmd
}

// NewRunner creates a new Runner instance and returns its pointer
func NewRunner() *Runner {
	return &Runner{
		cmds:  map[string]*exec.Cmd{},
		start: make(chan string),
		done:  make(chan struct{}),
	}
}

// Run initializes runner with given parameters.
func (r *Runner) Run() {
	for fileName := range r.start {
		cmd, err := runCommand(fileName)
		if err != nil {
			log.Printf("Could not run the go binary: %s \n", err)
			r.kill(cmd)

			continue
		}

		r.cmds[fileName] = cmd
		removeFile(fileName)

		go func(cmd *exec.Cmd) {
			if err := cmd.Wait(); err != nil {
				//log.Printf("process interrupted: %s \n", err)
				r.kill(cmd)
			}
		}(r.cmds[fileName])
	}
}

// Restart kills the process, removes the old binary and
// restarts the new process
func (r *Runner) restart(fileName string) {
	r.kill(r.cmds[fileName])

	r.start <- fileName
}

func (r *Runner) kill(cmd *exec.Cmd) {
	if cmd != nil {
		cmd.Process.Kill()
	}
}

func (r *Runner) Close() {
	close(r.start)
	for _, cmd := range r.cmds {
		r.kill(cmd)
	}
	close(r.done)
}

func (r *Runner) Wait() {
	<-r.done
}

// runCommand runs the command with given name and arguments. It copies the
// logs to standard output
func runCommand(name string, args ...string) (*exec.Cmd, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(name, args...)
	if tp := viper.GetString(fs.DebugKey); tp != "" {
		cmd.Dir = path.Join(currentPath, tp)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return cmd, err
	}

	if err := cmd.Start(); err != nil {
		return cmd, err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return cmd, nil
}

func removeFile(fileName string) {
	if fileName != "" {
		cmd := exec.Command("rm", fileName)
		cmd.Run()
		cmd.Wait()
	}
}
