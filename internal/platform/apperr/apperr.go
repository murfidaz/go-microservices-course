package apperr

import (
	"errors"
	"maps"
	"slices"
	"strings"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type ValidationError struct {
	Fields map[string]string
}

func (v *ValidationError) Add(field, problem string) {
	if v.Fields == nil {
		v.Fields = make(map[string]string)
	}
	v.Fields[field] = problem
}

func (v *ValidationError) Err() error {
	if len(v.Fields) == 0 {
		return nil
	}
	return v
}

func (v *ValidationError) Error() string {
	fields := slices.Sorted(maps.Keys(v.Fields))
	return "invalid input: " + strings.Join(fields, ", ")
}
