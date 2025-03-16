package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	clockTicks = 0
)

func Exec(dir, name string, arg ...string) (err error) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if dir != "" {
		cmd.Dir = dir
	}

	err = cmd.Run()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	return
}

func ExecInput(dir, input, name string, arg ...string) (err error) {
	cmd := exec.Command(name, arg...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err,
				"utils: Failed to get stdin in exec '%s'", name),
		}
		return
	}

	if dir != "" {
		cmd.Dir = dir
	}

	err = cmd.Start()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	var wrErr error
	go func() {
		defer func() {
			wrErr = stdin.Close()
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to close stdin in exec '%s'",
						name,
					),
				}
			}
		}()

		_, wrErr = io.WriteString(stdin, input)
		if wrErr != nil {
			wrErr = &errortypes.ExecError{
				errors.Wrapf(
					wrErr,
					"utils: Failed to write stdin in exec '%s'",
					name,
				),
			}
			return
		}
	}()

	err = cmd.Wait()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	if wrErr != nil {
		err = wrErr
		return
	}

	return
}

func ExecInputOutput(input, name string, arg ...string) (
	output string, err error) {

	cmd := exec.Command(name, arg...)

	stdout := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to get stdin in exec '%s'", name),
		}
		return
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	var wrErr error
	go func() {
		defer func() {
			wrErr = stdin.Close()
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to close stdin in exec '%s'",
						name,
					),
				}
			}
		}()

		_, wrErr = io.WriteString(stdin, input)
		if wrErr != nil {
			wrErr = &errortypes.ExecError{
				errors.Wrapf(
					wrErr,
					"utils: Failed to write stdin in exec '%s'",
					name,
				),
			}
			return
		}
	}()

	err = cmd.Wait()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	if wrErr != nil {
		err = wrErr
		return
	}

	output = string(stdout.Bytes())

	return
}

func ExecInputOutputCombindLogged(input, name string, arg ...string) (
	output string, err error) {

	cmd := exec.Command(name, arg...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to get stdin in exec '%s'", name),
		}
		return
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	var wrErr error
	go func() {
		defer func() {
			wrErr = stdin.Close()
			if wrErr != nil {
				wrErr = &errortypes.ExecError{
					errors.Wrapf(
						wrErr,
						"utils: Failed to close stdin in exec '%s'",
						name,
					),
				}
			}
		}()

		_, wrErr = io.WriteString(stdin, input)
		if wrErr != nil {
			wrErr = &errortypes.ExecError{
				errors.Wrapf(
					wrErr,
					"utils: Failed to write stdin in exec '%s'",
					name,
				),
			}
			return
		}
	}()

	err = cmd.Wait()

	output = stdout.String()
	errOutput := stderr.String()

	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output":       output,
			"error_output": errOutput,
			"cmd":          name,
			"arg":          arg,
			"error":        err,
		}).Error("utils: Process exec error")

		return
	}

	if wrErr != nil {
		logrus.WithFields(logrus.Fields{
			"output":       output,
			"error_output": errOutput,
			"cmd":          name,
			"arg":          arg,
			"error":        wrErr,
		}).Error("utils: Process exec error")

		return
	}

	output = string(stdout.Bytes())

	return
}

func ExecOutput(dir, name string, arg ...string) (output string, err error) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr

	if dir != "" {
		cmd.Dir = dir
	}

	outputByt, err := cmd.Output()
	if outputByt != nil {
		output = string(outputByt)
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	return
}

func ExecCombinedOutput(dir, name string, arg ...string) (
	output string, err error) {

	cmd := exec.Command(name, arg...)

	if dir != "" {
		cmd.Dir = dir
	}

	outputByt, err := cmd.CombinedOutput()
	if outputByt != nil {
		output = string(outputByt)
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}
		return
	}

	return
}

func ExecCombinedOutputLogged(ignores []string, name string, arg ...string) (
	output string, err error) {

	cmd := exec.Command(name, arg...)

	outputByt, err := cmd.CombinedOutput()
	if outputByt != nil {
		output = string(outputByt)
	}

	if err != nil && ignores != nil {
		for _, ignore := range ignores {
			if strings.Contains(output, ignore) {
				err = nil
				output = ""
				break
			}
		}
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output": output,
			"cmd":    name,
			"arg":    arg,
			"error":  err,
		}).Error("utils: Process exec error")
		return
	}

	return
}

func ExecCombinedOutputLoggedDir(ignores []string,
	dir, name string, arg ...string) (
	output string, err error) {

	cmd := exec.Command(name, arg...)
	if dir != "" {
		cmd.Dir = dir
	}

	outputByt, err := cmd.CombinedOutput()
	if outputByt != nil {
		output = string(outputByt)
	}

	if err != nil && ignores != nil {
		for _, ignore := range ignores {
			if strings.Contains(output, ignore) {
				err = nil
				output = ""
				break
			}
		}
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output": output,
			"cmd":    name,
			"arg":    arg,
			"error":  err,
		}).Error("utils: Process exec error")
		return
	}

	return
}

func ExecOutputLogged(ignores []string, name string, arg ...string) (
	output string, err error) {

	cmd := exec.Command(name, arg...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	output = stdout.String()
	errOutput := stderr.String()

	if err != nil && ignores != nil {
		for _, ignore := range ignores {
			if strings.Contains(output, ignore) ||
				strings.Contains(errOutput, ignore) {

				err = nil
				output = ""
				break
			}
		}
	}
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrapf(err, "utils: Failed to exec '%s'", name),
		}

		logrus.WithFields(logrus.Fields{
			"output":       output,
			"error_output": errOutput,
			"cmd":          name,
			"arg":          arg,
			"error":        err,
		}).Error("utils: Process exec error")
		return
	}

	return
}

func getClockTicks() (ticks int) {
	if clockTicks != 0 {
		ticks = clockTicks
		return
	}

	resp, err := commander.Exec(&commander.Opt{
		Name:    "getconf",
		Args:    []string{"CLK_TCK"},
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		ticks = 100
		clockTicks = 100
		return
	}

	if resp.Output != nil {
		ticks, _ = strconv.Atoi(strings.TrimSpace(string(resp.Output)))
	}

	if ticks == 0 {
		ticks = 100
		clockTicks = 100
		return
	}

	clockTicks = ticks
	return
}

func GetProcessTimestamp(pid int) (timestamp time.Time, err error) {
	procPath := filepath.Join("/proc", strconv.Itoa(pid))

	_, err = os.Stat(procPath)
	if os.IsNotExist(err) {
		err = nil
		return
	}

	statPath := filepath.Join(procPath, "stat")
	statData, err := os.ReadFile(statPath)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read process stat"),
		}
		return
	}

	statFields := strings.Fields(string(statData))
	if len(statFields) < 22 {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Invalid process state format"),
		}
		return
	}

	startTimeTicks, err := strconv.ParseInt(statFields[21], 10, 64)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to process stat"),
		}
		return
	}

	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to read uptime"),
		}
		return
	}

	uptimeFields := strings.Fields(string(uptimeData))
	if len(uptimeFields) < 1 {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Invalid uptime format"),
		}
		return
	}

	systemUptimeSec, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "utils: Failed to process uptime"),
		}
		return
	}

	processStartTimeSec := float64(startTimeTicks) / float64(getClockTicks())
	processUptimeSec := systemUptimeSec - processStartTimeSec
	timestamp = time.Now().Add(
		time.Duration(processUptimeSec * -float64(time.Second)))

	return
}
