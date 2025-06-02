package engine

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const shellEnvExport = `
echo "<STARTER_ENV_EXPORT>"
env
echo "</STARTER_ENV_EXPORT>"
`

var colorRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type BashEngine struct {
	cmd        *exec.Cmd
	cwd        string
	shell      string
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	curEnvKeys set.Set
	starter    *Engine
}

func (b *BashEngine) Init(strt *Engine) (err error) {
	b.starter = strt
	b.curEnvKeys = set.NewSet()

	_, err = exec.LookPath("bash")
	if err == nil {
		b.shell = "bash"
	} else {
		b.shell = "sh"
		err = nil
	}

	for _, pairStr := range os.Environ() {
		pair := strings.SplitN(pairStr, "=", 2)
		b.curEnvKeys.Add(pair[0])
	}

	return
}

func (b *BashEngine) streamOut(reader io.Reader) (env []string) {
	scanner := bufio.NewScanner(reader)
	envCapture := false

	for scanner.Scan() {
		line := scanner.Text()
		cleanLine := colorRe.ReplaceAllString(line, "")

		if envCapture {
			if strings.HasPrefix(line, "</STARTER_ENV_EXPORT>") {
				envCapture = false
			} else {
				env = append(env, line)
			}
		} else {
			if strings.HasPrefix(line, "<STARTER_ENV_EXPORT>") {
				envCapture = true
			} else if strings.HasPrefix(
				line, "+ echo '<STARTER_ENV_EXPORT>'") || strings.HasPrefix(
				line, "+ echo '</STARTER_ENV_EXPORT>'") {

			} else {
				if cleanLine != "" {
					b.starter.ProcessOutput(cleanLine)
				}
			}
		}
	}

	return
}

func (b *BashEngine) streamErr(reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		cleanLine := colorRe.ReplaceAllString(line, "")
		if cleanLine != "" {
			b.starter.ProcessOutput(cleanLine)
		}
	}
}

func (b *BashEngine) Run(block string) (err error) {
	cmd := exec.Command(b.shell, "-v", "-c", block+shellEnvExport)
	cmd.Env = b.starter.GetEnviron()
	cmd.Dir = b.starter.GetCwd()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to create stdout pipe"),
		}
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to create stderr pipe"),
		}
		return
	}

	err = cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to start command"),
		}
		return
	}

	env := []string{}
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		env = b.streamOut(stdout)
		wg.Done()
	}()

	go func() {
		b.streamErr(stderr)
		wg.Done()
	}()

	err = cmd.Wait()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Exit error in "+b.shell),
		}
		return
	}

	wg.Wait()

	for _, pairStr := range env {
		pair := strings.SplitN(pairStr, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key, val := pair[0], pair[1]

		if key == "PWD" {
			b.starter.UpdateCwd(val)
		} else if !b.curEnvKeys.Contains(key) {
			b.starter.UpdateEnv(key, val)
		}
	}

	return
}
