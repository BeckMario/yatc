package internal

import (
	"encoding/json"
)

func Ptr[T any](x T) *T {
	return &x
}

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{
		m: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(value T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
	s.m[value] = struct{}{}
}

func (s *Set[T]) Has(value T) bool {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
	_, ok := s.m[value]
	return ok
}

func (s *Set[T]) Remove(value T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
	delete(s.m, value)
}

func (s *Set[T]) ToArray() []T {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}

	keys := make([]T, len(s.m))

	i := 0
	for k := range s.m {
		keys[i] = k
		i++
	}

	return keys
}

func (s *Set[T]) MarshalJSON() ([]byte, error) {
	array := s.ToArray()
	return json.Marshal(&array)
}

func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var values []T
	err := json.Unmarshal(data, &values)
	if err != nil {
		return err
	}
	s.m = SetOf[T](values...).m
	return nil
}

func SetOf[T comparable](values ...T) *Set[T] {
	s := NewSet[T]()
	for _, value := range values {
		s.Add(value)
	}
	return &s
}

type contextKey int

const (
	ContextKeyTraceParent contextKey = iota
)
