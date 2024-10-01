package eval

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

func ValidateEval(statement string) (err error) {
	if len(statement) == 0 {
		err = &errortypes.ParseError{
			errors.New("eval: Empty statement"),
		}
		return
	}

	if len(statement) > StatementMaxLength {
		err = &errortypes.ParseError{
			errors.Newf("eval: Statement exceeds max length"),
		}
		return
	}

	for i, c := range statement {
		if !StatementSafeCharacters.Contains(c) {
			err = &errortypes.ParseError{
				errors.Newf("eval: Illegal char (%s) at %d", string(c), i+1),
			}
			return
		}
	}

	return
}
