package engine

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/utils"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/tools/logger"
)

type Engine struct {
	cwd        string
	env        map[string]string
	blocks     []*Block
	bash       *BashEngine
	python     *PythonEngine
	lock       sync.Mutex
	outputLock sync.Mutex
	startPhase string
	queue      chan []*Block
}

func (e *Engine) UpdateEnv(key, val string) {
	e.env[key] = val
}

func (e *Engine) UpdateCwd(cwd string) {
	e.cwd = cwd
}

func (e *Engine) ProcessOutput(output string) {
	e.outputLock.Lock()
	fmt.Println(output)
	e.outputLock.Unlock()
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

func (e *Engine) Init(phase string) (err error) {
	e.startPhase = phase
	e.queue = make(chan []*Block, QueueSize)

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

	go e.runner()

	return
}

func (e *Engine) getBlocks() (blocks []*Block) {
	blocks = <-e.queue

	for {
		select {
		case req := <-e.queue:
			blocks = req
		default:
			return blocks
		}
	}
}

func (e *Engine) UpdateSpec(data string) (err error) {
	blocks, err := Parse(data)
	if err != nil {
		return
	}

	e.blocks = blocks

	return
}

func (e *Engine) runner() {
	initialized := false

	for {
		blocks := e.getBlocks()

		var phase string
		if !initialized {
			phase = e.startPhase
		} else {
			phase = Reload
		}

		err := e.run(phase, blocks)
		if err != nil {
			if initialized {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to run spec")
			} else {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to run initial spec")
				utils.DelayExit(1, 1*time.Second)
			}
		}

		initialized = true
	}
}

func (e *Engine) run(phase string, blocks []*Block) (err error) {
	for _, block := range blocks {
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
		}
	}

	return
}

func (e *Engine) Queue(data string) {
	blocks, err := Parse(data)
	if err != nil {
		return
	}

	e.lock.Lock()
	if len(e.queue) >= QueueSize-16 {
		return
	}
	e.queue <- blocks
	e.lock.Unlock()
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
