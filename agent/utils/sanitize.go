package utils

import (
	"strings"

	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/tools/commander"
)

type fileOp struct {
	cmd  string
	args []string
	path string
}

type findOp struct {
	path    string
	pattern string
	action  string
}

func Sanitize() (err error) {
	operations := []fileOp{
		{"rm", []string{"-f"}, "/var/lib/systemd/random-seed"},
		{"rm", []string{"-f"}, "/etc/machine-id"},
		{"rm", []string{"-rf"}, "/root/.cache"},
		{"shred", []string{"-u"}, "/root/.ssh/authorized_keys"},
		{"shred", []string{"-u"}, "/root/.bash_history"},
		{"shred", []string{"-u"}, "/var/log/lastlog"},
		{"shred", []string{"-u"}, "/var/log/secure"},
		{"shred", []string{"-u"}, "/var/log/utmp"},
		{"shred", []string{"-u"}, "/var/log/wtmp"},
		{"shred", []string{"-u"}, "/var/log/btmp"},
		{"shred", []string{"-u"}, "/var/log/dmesg"},
		{"shred", []string{"-u"}, "/var/log/dmesg.old"},
		{"shred", []string{"-u"}, "/var/lib/systemd/random-seed"},
	}

	findOps := []findOp{
		{"/var/tmp", "-name dnf-*", "rm -rf"},
		{"/home", "-type d -name .cache", "rm -rf"},
		{"/home", "-type f -name .bash_history", "rm -f"},
		{"/var/log", "-type f -name *.gz", "rm -f"},
		{"/var/log", "-type f -name *.[0-9]", "rm -f"},
		{"/var/log", "-type f -name *-????????", "rm -f"},
		{"/var/lib/cloud/instances", "-mindepth 1", "rm -rf"},
		{"/etc/ssh", "-type f -name *_key", "shred -u"},
		{"/etc/ssh", "-type f -name *_key.pub", "shred -u"},
		{"/var/log", "-mtime -1 -type f", "truncate -s 0"},
	}

	_, _ = commander.Exec(&commander.Opt{
		Name:    "sync",
		Args:    []string{},
		PipeOut: true,
		PipeErr: true,
	})

	for _, op := range operations {
		_, _ = commander.Exec(&commander.Opt{
			Name:    op.cmd,
			Args:    append(op.args, op.path),
			PipeOut: true,
			PipeErr: true,
		})
	}

	for _, op := range findOps {
		args := []string{op.path}
		if op.pattern != "" {
			args = append(args, strings.Split(op.pattern, " ")...)
		}
		args = append(args, "-exec")
		args = append(args, strings.Split(op.action, " ")...)
		args = append(args, "{}", ";")

		_, _ = commander.Exec(&commander.Opt{
			Name:    "find",
			Args:    args,
			PipeOut: true,
			PipeErr: true,
		})
	}

	_, _ = commander.Exec(&commander.Opt{
		Name:    "touch",
		Args:    []string{"/etc/machine-id"},
		PipeOut: true,
		PipeErr: true,
	})

	_, _ = commander.Exec(&commander.Opt{
		Name:    "sync",
		Args:    []string{},
		PipeOut: true,
		PipeErr: true,
	})

	_, _ = commander.Exec(&commander.Opt{
		Name:    "history",
		Args:    []string{"-c"},
		PipeOut: true,
		PipeErr: true,
	})

	_, _ = commander.Exec(&commander.Opt{
		Name:    "fstrim",
		Args:    []string{"-av"},
		PipeOut: true,
		PipeErr: true,
	})

	return
}

func SanitizeImds() (err error) {
	_, _ = commander.Exec(&commander.Opt{
		Name:    "shred",
		Args:    []string{"-u", constants.ImdsLogPath},
		PipeOut: true,
		PipeErr: true,
	})

	_, _ = commander.Exec(&commander.Opt{
		Name:    "shred",
		Args:    []string{"-u", constants.ImdsConfPath},
		PipeOut: true,
		PipeErr: true,
	})

	return
}
