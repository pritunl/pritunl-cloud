package engine

import (
	"fmt"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Engine struct {
	cwd    string
	env    map[string]string
	blocks []*Block
	bash   *BashEngine
	python *PythonEngine
}

func (e *Engine) UpdateEnv(key, val string) {
	e.env[key] = val
}

func (e *Engine) UpdateCwd(cwd string) {
	e.cwd = cwd
}

func (e *Engine) ProcessOutput(output string) {
	fmt.Println(output)
}

func (e *Engine) GetEnv() map[string]string {
	return e.env
}

func (e *Engine) GetCwd() string {
	return e.cwd
}

func (e *Engine) GetEnviron() (env []string) {
	env = []string{}

	for _, pair := range os.Environ() {
		key := strings.SplitN(pair, "=", 2)[0]
		if e.env[key] == "" {
			env = append(env, pair)
		}
	}

	for key, val := range e.env {
		env = append(env, key+"="+val)
	}

	return
}

func (e *Engine) Init() (err error) {
	e.cwd, err = os.Getwd()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "starter: Failed to get working dir"),
		}
		return
	}

	e.env = map[string]string{}

	e.bash = &BashEngine{}
	err = e.bash.Init(e)
	if err != nil {
		return
	}

	e.python = &PythonEngine{}
	err = e.python.Init(e)
	if err != nil {
		return
	}
	defer func() {
		err2 := e.python.Exit()
		if err2 != nil {
			panic(err2)
		}
	}()

	data, err := utils.ReadLines("/etc/pritunl-deploy.md")
	if err != nil {
		return
	}

	err = e.UpdateSpec(strings.Join(data, "\n"))
	if err != nil {
		return
	}

	return
}

func (e *Engine) UpdateSpec(data string) (err error) {
	blocks, err := Parse(data)
	if err != nil {
		return
	}

	e.blocks = blocks

	return
}

func (e *Engine) Run(phase string) (err error) {
	for _, block := range e.blocks {
		switch phase {
		case Initial:
			break
		case Reboot:
			if block.Phase != Reboot && block.Phase != Reload {
				continue
			}
			break
		case Reload:
			if block.Phase != Reload {
				continue
			}
			break
		}

		if block.Type == "shell" {
			err = e.RunBash(block.Code)
			if err != nil {
				return
			}
		} else if block.Type == "python" {
			err = e.RunPython(block.Code)
			if err != nil {
				return
			}
		} else {
			err = &errortypes.ParseError{
				errors.Newf("starter: Unknown block type %s", block.Type),
			}
			return
		}
	}

	return
}

func (e *Engine) RunBash(block string) (err error) {
	err = e.bash.Run(block)
	if err != nil {
		return
	}

	return
}

func (e *Engine) RunPython(block string) (err error) {
	err = e.python.Run(block)
	if err != nil {
		return
	}

	return
}
