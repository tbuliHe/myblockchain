package core

import "fmt"

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

func (s *State) Put(key, value []byte) error {
	s.data[string(key)] = value
	return nil
}

func (s *State) Delete(key string) error {
	delete(s.data, string(key))
	return nil
}

func (s *State) Get(k []byte) ([]byte, error) {
	key := string(k)
	value, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("given key %s not found", key)
	}
	return value, nil
}
