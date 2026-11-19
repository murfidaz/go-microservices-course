package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Env membaca typed values dari variable env, pembacaan tidak berhenti ketika ada kesalahan, dia akan mengumpulkan semuanya terlebih dahulu, jadi jika ada kesalahan dalam config akan dikumpulkan dalam satu kali proses.
type Env struct {
	lookup   func(key string) (string, bool)
	problems []string
}

func New(lookup func(key string) (string, bool)) *Env {
	return &Env{lookup: lookup}
}

// ambil nilai sebuah variable tanpa spasi
func (e *Env) get(key string) (string, bool) {
	v, _ := e.lookup(key)
	v = strings.TrimSpace(v)
	return v, v != ""
}

func (e *Env) Invalid(key, problem string) {
	e.problems = append(e.problems, key+": "+problem)
}

func (e *Env) Err() error {
	if len(e.problems) == 0 {
		return nil
	}

	return errors.New("invalid configurations:\n " + strings.Join(e.problems, "\n "))
}

func (e *Env) String(key, def string) string {
	if v, ok := e.get(key); ok {
		return v
	}
	return def
}

func (e *Env) Required(key string) string {
	v, ok := e.get(key)
	if !ok {
		e.Invalid(key, "required, but not set")
	}
	return v
}

func (e *Env) OneOf(key, def string, allowed ...string) string {
	v := e.String(key, def)
	if !slices.Contains(allowed, v) {
		e.Invalid(key, fmt.Sprintf("%q is not one of %s", v, strings.Join(allowed, ", ")))
	}
	return v
}

func Parse[T any](e *Env, key string, def T, parse func(string) (T, error)) T {
	v, ok := e.get(key)
	if !ok {
		return def
	}
	x, err := parse(v)
	if err != nil {
		e.Invalid(key, fmt.Sprintf("invalid value %q", v))
		return def
	}
	return x
}

func (e *Env) Int(key string, def int) int {
	return Parse(e, key, def, strconv.Atoi)
}

func (e *Env) Duration(key string, def time.Duration) time.Duration {
	return Parse(e, key, def, time.ParseDuration)
}
