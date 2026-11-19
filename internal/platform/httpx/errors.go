package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/murfidaz/go-microservices-course/internal/platform/apperr"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func classify(err error) (int, string) {
	var reqErr *requestError
	var valError *apperr.ValidationError
	switch {
	case errors.As(err, &reqErr):
		return reqErr.status, reqErr.code
	case errors.As(err, &valError):
		return http.StatusUnprocessableEntity, "invalid_input"
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict, "conflict"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	status, code := classify(err)
	d := errorDetail{Code: code, Message: err.Error()}
	if valErr, ok := errors.AsType[*apperr.ValidationError](err); ok {
		d.Fields = valErr.Fields
	}

	if status >= 500 {
		slog.ErrorContext(r.Context(), "request failed", "err", err)
		d.Message = "internal server error"
	}
	writeError(w, r, status, d)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, d errorDetail) {
	d.RequestID = RequestIDFrom(r.Context())
	body, _ := json.Marshal(errorBody{Error: d})
	writeBody(w, status, body)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	Error(w, r, fmt.Errorf("route %w", apperr.ErrNotFound))
}
