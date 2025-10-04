package types

import "github.com/pritunl/pritunl-cloud/spec"

type Journal struct {
	Index int32  `json:"index"`
	Key   string `json:"key"`
	Type  string `json:"type"`
	Unit  string `json:"unit"`
	Path  string `json:"path"`
}

func NewJournals(spc *spec.Spec) []*Journal {
	if spc == nil || spc.Journal == nil {
		return nil
	}

	jrnls := []*Journal{}
	for _, jrnl := range spc.Journal.Inputs {
		jrnls = append(jrnls, &Journal{
			Index: jrnl.Index,
			Key:   jrnl.Key,
			Type:  jrnl.Type,
			Unit:  jrnl.Unit,
			Path:  jrnl.Path,
		})
	}

	return jrnls
}
