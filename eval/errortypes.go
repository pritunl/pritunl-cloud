package eval

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

type EvalError struct {
	Statement string
	Index     int
	ErrIndex  int
	Length    int
	errors.DropboxError
}

func NewEvalError(statement string, index, errorIndex int,
	length int, templMsg string, args ...interface{}) (err error) {

	evalErr := &EvalError{
		Statement: statement,
		Index:     index + 1,
		ErrIndex:  errorIndex + 1,
		Length:    length,
	}

	tmpl, err := template.New("eval").Parse(templMsg)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "eval: Failed to parse eval error template"),
		}
		return
	}

	errorMsg := &bytes.Buffer{}
	err = tmpl.Execute(errorMsg, evalErr)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "eval: Failed to execute eval error template"),
		}
		return
	}

	errorMsgStr := fmt.Sprintf(errorMsg.String(), args...)
	errorMsgStr += fmt.Sprintf(" index=%d", evalErr.Index)
	errorMsgStr += fmt.Sprintf(" error_index=%d", evalErr.ErrIndex)
	errorMsgStr += fmt.Sprintf(" statement=\"%s\"", evalErr.Statement)

	evalErr.DropboxError = errors.New(errorMsgStr)
	err = evalErr

	return
}
