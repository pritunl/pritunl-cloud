package plan

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/eval"
)

type Data struct {
	Unit     Unit     `json:"unit"`
	Instance Instance `json:"instance"`
}

type Unit struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Instance struct {
	Name          string `json:"name"`
	State         string `json:"state"`
	Action        string `json:"action"`
	Processors    int    `json:"processors"`
	Memory        int    `json:"memory"`
	LastTimestamp int    `json:"last_timestamp"`
	LastHeartbeat int    `json:"last_heartbeat"`
}

func (d *Data) Export() (data eval.Data, err error) {
	dataByt, err := json.Marshal(d)
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

func GetEmtpyData() (data eval.Data, err error) {
	dataStrct := Data{}

	data, err = dataStrct.Export()
	if err != nil {
		return
	}

	return
}
