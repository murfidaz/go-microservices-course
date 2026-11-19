package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxBodyBytes = 1 << 20 //1 MiB

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		Error(w, r, fmt.Errorf("encode response: %w", err))
		return
	}
	writeBody(w, status, body)
}

func writeBody(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(body, '\n'))
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return badRequest("body must contain a single JSON value")
	}
	return nil
}

type requestError struct {
	status int
	code   string
	msg    string
}

func (e *requestError) Error() string { return e.msg }

func badRequest(msg string) error {
	return &requestError{http.StatusBadRequest, "bad_request", msg}
}

func decodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var sizeErr *http.MaxBytesError
	switch {
	case errors.Is(err, io.EOF):
		return badRequest("body must not be empty")
	case errors.As(err, &syntaxErr):
		return badRequest(fmt.Sprintf("malformed JSON at byte %d", syntaxErr.Offset))
	case errors.Is(err, io.ErrUnexpectedEOF):
		return badRequest("malformed JSON")
	case errors.As(err, &typeErr):
		return badRequest(fmt.Sprintf("field %q must be of type %s", typeErr.Field, typeErr.Type))
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		return badRequest(strings.TrimPrefix(err.Error(), "json: "))
	case errors.As(err, &sizeErr):
		msg := fmt.Sprintf("body must be at most %d bytes", sizeErr.Limit)
		return &requestError{http.StatusRequestEntityTooLarge, "body_too_large", msg}
	default:
		return badRequest("invalid JSON body")
	}
}
