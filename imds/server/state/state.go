package state

import (
	"sync"

	"github.com/pritunl/pritunl-cloud/imds/types"
)

var Global = &Store{
	State:    &types.State{},
	output:   make(chan *types.Entry, 10000),
	journals: map[string]chan *types.Entry{},
}

type Store struct {
	State    *types.State
	output   chan *types.Entry
	journals map[string]chan *types.Entry
	lock     sync.RWMutex
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

func (s *Store) AppendJournalOutput(key string, entry *types.Entry) {
	s.lock.Lock()
	output, exists := s.journals[key]
	if !exists {
		output = make(chan *types.Entry, 10000)
		s.journals[key] = output
	}
	s.lock.Unlock()

	if len(output) > 9000 {
		return
	}
	output <- entry
}

func (s *Store) GetJournals() (journals map[string][]*types.Entry) {
	journals = map[string][]*types.Entry{}

	s.lock.RLock()
	keys := make([]string, 0, len(s.journals))
	outputs := make(map[string]chan *types.Entry)
	for key, output := range s.journals {
		keys = append(keys, key)
		outputs[key] = output
	}
	s.lock.RUnlock()

	for _, key := range keys {
		output := outputs[key]
		if output == nil {
			continue
		}

		var entries []*types.Entry
		for {
			select {
			case entry := <-output:
				entries = append(entries, entry)
			default:
				if len(entries) > 0 {
					journals[key] = entries
				}
				goto nextKey
			}
		}
	nextKey:
	}

	return
}

func Init() (err error) {
	return
}
