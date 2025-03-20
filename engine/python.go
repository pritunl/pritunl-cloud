package engine

import (
	"bufio"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	PythonExec = "python3"
)

const pyEngine = `#!/usr/bin/env python3
import platform
import os
import json

def _pystarter_get_startup():
    python_version = platform.python_version()
    compiler = platform.python_compiler()
    os_name = platform.system().lower() + " " + platform.release()
    build_date = " ".join(platform.python_build()[1].split()[1:4])
    return f"Python {python_version} ({build_date}) [{compiler}] on {os_name}"

def export(name, arg):
    name = ''.join(c for c in name if c.isalnum() or c == '_')
    arg = json.dumps(arg)
    print(f"<PYSTARTER_EXPORT_VAR>{name}={arg}</PYSTARTER_EXPORT_VAR>")

print(_pystarter_get_startup())
print("<PYSTARTER_INIT_COMPLETE/>")

while True:
    data = input()
    if not data:
        continue

    data = json.loads(data)

    if data["type"] == "exit":
        exit(0)
    elif data["type"] == "env":
        for key, val in data["env"].items():
            os.environ[key] = val
        print("<PYSTARTER_UPDATE_COMPLETE/>")
    elif data["type"] == "chdir":
        os.chdir(data["input"])
        print("<PYSTARTER_UPDATE_COMPLETE/>")
    elif data["type"] == "exec":
        exec(data["input"])
        print(f"<PYSTARTER_EXEC_COMPLETE>"
            f"{os.getcwd()}</PYSTARTER_EXEC_COMPLETE>")
`

type pythonData struct {
	Type  string            `json:"type"`
	Input string            `json:"input"`
	Env   map[string]string `json:"env"`
}

type PythonEngine struct {
	cmd         *exec.Cmd
	cwd         string
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	cmdErr      error
	inputLines  *list.List
	waiter      chan bool
	initialized bool
	starter     *Engine

	output string
}

func (p *PythonEngine) Init(strt *Engine) (err error) {
	p.starter = strt

	return
}

func (p *PythonEngine) start() (err error) {
	p.waiter = make(chan bool, 16)
	p.inputLines = list.New()

	p.cmd = exec.Command(PythonExec, "-u", "-c", pyEngine)
	p.cmd.Env = p.starter.GetEnviron()
	p.cmd.Dir = p.starter.GetCwd()

	p.stdout, err = p.cmd.StdoutPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to get py stdout"),
		}
		return
	}

	p.stderr, err = p.cmd.StderrPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to get py stderr"),
		}
		return
	}

	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to get py stdin"),
		}
		return
	}

	err = p.cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to start py"),
		}
		return
	}
	p.wait()

	p.copyOutput(p.stdout)
	p.copyOutput(p.stderr)

	<-p.waiter

	return
}

func (p *PythonEngine) flushOutput() {
	if p.output == "" {
		return
	}
	output := p.output
	p.output = ""
	output = strings.TrimRight(output, "\n")
	p.starter.ProcessOutput(output)
}

func (p *PythonEngine) copyOutput(src io.ReadCloser) {
	go func() {
		scanner := bufio.NewScanner(src)
		for scanner.Scan() {
			output := scanner.Text()

			if !p.initialized {
				if strings.Contains(output, "<PYSTARTER_INIT_COMPLETE/>") {
					p.initialized = true
					p.waiter <- true
				} else {
					p.starter.ProcessOutput(output)
				}

				continue
			}

			execDoneStart := strings.Index(output, "<PYSTARTER_EXEC_COMPLETE>")
			exportStart := strings.Index(output, "<PYSTARTER_EXPORT_VAR>")

			if execDoneStart != -1 {
				execDoneStart += 25
				execDoneEnd := strings.Index(output,
					"</PYSTARTER_EXEC_COMPLETE>")
				if execDoneEnd == -1 {
					err := &errortypes.ExecError{
						errors.Newf(
							"starter: Incomplete exec response '%s'", output),
					}
					panic(err)
				}

				p.starter.UpdateCwd(output[execDoneStart:execDoneEnd])
				p.waiter <- true
			} else if exportStart != -1 {
				exportStart += 22
				exportEnd := strings.Index(output, "</PYSTARTER_EXPORT_VAR>")
				if exportEnd == -1 {
					err := &errortypes.ExecError{
						errors.Newf(
							"starter: Incomplete export '%s'", output),
					}
					panic(err)
				}

				pair := strings.SplitN(output[exportStart:exportEnd], "=", 2)
				if len(pair) == 2 {
					key, val := pair[0], pair[1]

					if val[0] == '"' && val[len(val)-1] == '"' {
						val = val[1 : len(val)-1]
					}

					p.starter.UpdateEnv(key, val)
				}
			} else if strings.Contains(output,
				"<PYSTARTER_UPDATE_COMPLETE/>") {

				p.waiter <- true
			} else {
				p.output += output + "\n"
			}
		}
	}()
}

func (p *PythonEngine) wait() {
	go func() {
		defer func() {
			if p.stdin != nil {
				_ = p.stdin.Close()
			}
		}()

		err := p.cmd.Wait()
		if err != nil {
			p.flushOutput()
			err = &errortypes.ExecError{
				errors.Wrap(err, "starter: Exit error in py"),
			}
			p.cmdErr = err
			p.waiter <- true
			return
		}

		p.waiter <- true
	}()
}

func (p *PythonEngine) updateEnv() (err error) {
	defer func() {
		p.flushOutput()
	}()
	p.waiter = make(chan bool, 16)

	data := &pythonData{
		Type: "env",
		Env:  p.starter.GetEnv(),
	}

	dataIn, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "starter: Failed to marshal env"),
		}
		return
	}

	_, err = fmt.Fprintln(p.stdin, string(dataIn))
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to update env in py"),
		}
		return
	}

	<-p.waiter

	return
}

func (p *PythonEngine) updateCwd() (err error) {
	defer func() {
		p.flushOutput()
	}()
	p.waiter = make(chan bool, 16)

	data := &pythonData{
		Type:  "chdir",
		Input: p.starter.GetCwd(),
	}

	dataIn, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "starter: Failed to marshal env"),
		}
		return
	}

	_, err = fmt.Fprintln(p.stdin, string(dataIn))
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to update env in py"),
		}
		return
	}

	<-p.waiter

	return
}

func (p *PythonEngine) run(code string) (err error) {
	defer func() {
		p.flushOutput()
	}()
	p.waiter = make(chan bool, 16)

	data := &pythonData{
		Type:  "exec",
		Input: code,
	}

	dataIn, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "starter: Failed to marshal code"),
		}
		return
	}

	_, err = fmt.Fprintln(p.stdin, string(dataIn))
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to run code in py"),
		}
		return
	}

	<-p.waiter

	return
}

func (p *PythonEngine) Exit() (err error) {
	p.waiter = make(chan bool, 16)

	if p.cmd == nil {
		return
	}

	data := &pythonData{
		Type: "exit",
	}

	dataIn, err := json.Marshal(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "starter: Failed to marshal exit signal"),
		}
		return
	}

	_, err = fmt.Fprintln(p.stdin, string(dataIn))
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "starter: Failed to run exit signal in py"),
		}
		return
	}

	<-p.waiter

	if p.cmdErr != nil {
		err = p.cmdErr
	}

	return
}

func (p *PythonEngine) Run(code string) (err error) {
	if p.cmd == nil {
		err = p.start()
		if err != nil {
			return
		}
	}

	err = p.updateEnv()
	if err != nil {
		return
	}

	err = p.updateCwd()
	if err != nil {
		return
	}

	for _, line := range strings.Split(code, "\n") {
		p.starter.ProcessOutput(">>> " + line)
	}

	err = p.run(code)
	if err != nil {
		return
	}

	if p.cmdErr != nil {
		err = p.cmdErr
		return
	}

	return
}
