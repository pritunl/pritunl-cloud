package prompt

import (
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/manifoldco/promptui"
	"github.com/pritunl/pritunl-cloud/errortypes"
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

func Confirm(label string) (resp bool, err error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			_, e := parseYesNo(input)
			if e != nil {
				return e
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "prompt: Prompt run error"),
		}
		return
	}

	resp, err = parseYesNo(result)
	if err != nil {
		return
	}

	return
}

func ConfirmDefault(label string, def bool) (resp bool, err error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if input == "" {
				return nil
			}

			_, e := parseYesNo(input)
			if e != nil {
				return e
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "prompt: Prompt run error"),
		}
		return
	}

	if result == "" {
		resp = def
		return
	}

	resp, err = parseYesNo(result)
	if err != nil {
		return
	}

	return
}
