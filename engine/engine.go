package engine

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
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
	fault      atomic.Value
	queue      chan []*Block
	OnStatus   func(status string)
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

func (e *Engine) Init() (err error) {
	e.queue = make(chan []*Block, QueueSize)
	e.fault.Store(false)

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

	return
}

func (e *Engine) StartRunner() {
	go e.runner()
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
	for {
		blocks := e.getBlocks()

		if !e.fault.Load().(bool) {
			e.OnStatus(types.ReloadingClean)
		} else {
			e.OnStatus(types.ReloadingFault)
		}

		_, err := e.Run(Reload, blocks)
		if err != nil {
			logger.WithFields(logger.Fields{
				"error": err,
			}).Error("agent: Failed to run spec")
			e.OnStatus(types.Fault)
		} else {
			e.OnStatus(types.Running)
		}
	}
}

func (e *Engine) Run(phase string, blocks []*Block) (fatal bool, err error) {
	for i, block := range blocks {
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

		err = e.runBlock(block.Type, block.Code)
		if err != nil {
			for _, block := range blocks[i:] {
				switch phase {
				case Initial:
					if block.Phase != Reload {
						fatal = true
					}
				case Reboot:
					if block.Phase != Reboot && block.Phase != Reload {
						continue
					}
					if block.Phase != Reload {
						fatal = true
					}
				case Reload:
					if block.Phase != Reload {
						continue
					}
				}
			}

			e.fault.Store(true)
			return
		}
	}

	e.fault.Store(false)

	return
}

func (e *Engine) runBlock(blockType, block string) (err error) {
	switch blockType {
	case "shell":
		err = e.bash.Run(block)
		if err != nil {
			return
		}
	case "python":
		err = e.python.Run(block)
		if err != nil {
			return
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
