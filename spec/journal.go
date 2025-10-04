package spec

import (
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Journal struct {
	Inputs []*Input `bson:"inputs" json:"inputs"`
}

type Input struct {
	Index int32  `bson:"index" json:"index"`
	Key   string `bson:"key" json:"key"`
	Type  string `bson:"type" json:"type"`
	Unit  string `bson:"unit" json:"unit"`
	Path  string `bson:"path" json:"path"`
}

func (j *Journal) Validate() (errData *errortypes.ErrorData, err error) {
	for _, input := range j.Inputs {
		if input.Key == "" {
			errData = &errortypes.ErrorData{
				Error:   "journal_key_missing",
				Message: "Missing journal key",
			}
			return
		}
		key := utils.FilterName(input.Key)
		if input.Key != key {
			errData = &errortypes.ErrorData{
				Error:   "journal_key_invalid",
				Message: "Journal key invalid",
			}
			return
		}
		input.Key = key

		switch input.Type {
		case Systemd:
			input.Path = ""
			if input.Unit == "" {
				errData = &errortypes.ErrorData{
					Error:   "systemd_unit_missing",
					Message: "Missing systemd unit",
				}
				return
			}
			inputUnit := utils.FilterUnit(input.Unit)
			if input.Unit != inputUnit {
				errData = &errortypes.ErrorData{
					Error:   "systemd_unit_invalid",
					Message: "Invalid systemd unit",
				}
				return
			}
			input.Unit = inputUnit
			break
		case File:
			input.Unit = ""
			if input.Path == "" {
				errData = &errortypes.ErrorData{
					Error:   "log_path_missing",
					Message: "Missing log path",
				}
				return
			}
			input.Path = utils.FilterPath(input.Path)
			if input.Path == "" {
				errData = &errortypes.ErrorData{
					Error:   "log_path_invalid",
					Message: "Invalid log path",
				}
				return
			}
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "unknown_input_type",
				Message: "Unknown input type",
			}
			return
		}
	}

	return
}

type JournalYaml struct {
	Name   string             `yaml:"name"`
	Kind   string             `yaml:"kind"`
	Inputs []JournalYamlInput `yaml:"inputs"`
}

type JournalYamlInput struct {
	Key  string `yaml:"key"`
	Type string `yaml:"type"`
	Unit string `yaml:"unit,omitempty"`
	Path string `yaml:"path,omitempty"`
}
