package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	AssumeYes = false
)

func parseYesNo(input string) (val bool, err error) {
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		val = true
		return
	} else if input == "n" || input == "no" {
		val = false
		return
	}

	err = &errortypes.ParseError{
		errors.New("prompt: Invalid confirm input"),
	}
	return
}

func ConfirmDefault(label string, def bool) (resp bool, err error) {
	if AssumeYes {
		resp = true
		return
	}

	var prompt string
	if def {
		prompt = fmt.Sprintf("%s [Y/n]: ", label)
	} else {
		prompt = fmt.Sprintf("%s [y/N]: ", label)
	}

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	input = strings.TrimSpace(input)

	if input == "" {
		resp = def
		return
	}

	resp, err = parseYesNo(input)
	if err != nil {
		return
	}

	return
}
