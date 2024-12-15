package state

import (
	"github.com/pritunl/pritunl-cloud/imds/types"
)

var Global = &Store{
	State:  &types.State{},
	output: make(chan *types.Entry, 10000),
}

type Store struct {
	State  *types.State
	output chan *types.Entry
}

func (s *Store) AppendOutput(entry *types.Entry) {
	if len(s.output) > 9000 {
		return
	}
	s.output <- entry
}

func (s *Store) GetOutput() (entries []*types.Entry) {
	for {
		select {
		case entry := <-s.output:
			entries = append(entries, entry)
		default:
			return
		}
	}
}

func Init() (err error) {
	return
}
