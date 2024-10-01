package plan

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/eval"
)

type Data struct {
	Instance Instance `json:"instance"`
}

type Instance struct {
	State      string `json:"state"`
	VirtState  string `json:"virt_state"`
	Processors int    `json:"processors"`
	Memory     int    `json:"memory"`
}

func GetEmtpyData() (data eval.Data, err error) {
	dataStrct := Data{}

	dataByt, err := json.Marshal(dataStrct)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "plan: Failed to marshal"),
		}
		return
	}

	data = eval.Data{}

	err = json.Unmarshal(dataByt, &data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "plan: Failed to unmarshal"),
		}
		return
	}

	return
}
