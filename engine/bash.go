package engine

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

const shellEnvExport = `
echo "<STARTER_ENV_EXPORT>"
env
echo "</STARTER_ENV_EXPORT>"
`

type BashEngine struct {
	cmd        *exec.Cmd
	cwd        string
	shell      string
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	curEnvKeys set.Set
	env        []string
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

func (b *BashEngine) Run(block string) (err error) {
	cmd := exec.Command(b.shell, "-c", block+shellEnvExport)
	cmd.Env = b.starter.GetEnviron()
	cmd.Dir = b.starter.GetCwd()

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Exit error in "+b.shell),
		}
		return
	}

	envBlock := false
	for _, line := range strings.Split(string(output), "\n") {
		if envBlock {
			if strings.HasPrefix(line, "</STARTER_ENV_EXPORT>") {
				envBlock = false
			} else {
				b.env = append(b.env, line)
			}
		} else {
			if strings.HasPrefix(line, "<STARTER_ENV_EXPORT>") {
				envBlock = true
			} else {
				if line != "" {
					b.starter.ProcessOutput(line)
				}
			}
		}
	}

	for _, pairStr := range b.env {
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
