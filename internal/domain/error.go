package domain

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

// Domain errors will be identified both by an
// error message and its associated error code
type (
	ErrorCode uint
	ErrorMsg  string
)

const (
	// Handled exceptions
	// Validation errors
	ValidationError ErrorMsg = "ValidationError"
	// Authorization errors
	UnauthorizedError ErrorMsg = "UnauthorizedError"
	// PostgreSQL errors
	NonExistentTableError  ErrorMsg = "NonExistentPostgreSQLTableError"
	NonExistentColumnError ErrorMsg = "NonExistentPostgreSQLColumnError"
	DuplicateKeyError      ErrorMsg = "DuplicateKeyPostgreSQLError"
	ForeignKeyError        ErrorMsg = "ForeignKeyPostgreSQLError"
	NotFoundError          ErrorMsg = "NotFoundPostgreSQLError"
	UnknownError           ErrorMsg = "UnkwnownPostgreSQLError"
)

var ErrorCodes = map[ErrorMsg]ErrorCode{
	// Validation
	ValidationError: 2400,
	// Authorization
	UnauthorizedError: 2401,
	// PostgresQL
	NonExistentTableError:  2501,
	NonExistentColumnError: 2502,
	DuplicateKeyError:      2503,
	ForeignKeyError:        2504,
	NotFoundError:          2506,
	UnknownError:           2507,
}

func getErrorCode(msg ErrorMsg) ErrorCode {
	return ErrorCodes[msg]
}

func getErrorMessage(err error) (ErrorMsg, string) {
	e := strings.Split(err.Error(), ":")
	details := strings.TrimSpace(strings.Join(e[1:], ""))
	return ErrorMsg(e[0]), details
}

// Error is the domain error entity that is
// returned and rendered when something fails
type Error struct {
	HTTPStatusCode int       `json:"-"`
	ErrorCode      ErrorCode `json:"code,omitempty"`
	ErrorText      string    `json:"message,omitempty"`
}

// Required to be able to render it as JSON response
func (e *Error) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// Bad request error
func ErrBadRequest(err error) render.Renderer {
	eMsg, details := getErrorMessage(err)
	errorCode := getErrorCode(eMsg)

	return &Error{
		HTTPStatusCode: http.StatusBadRequest,
		ErrorCode:      errorCode,
		ErrorText:      details,
	}
}

// Not found error
func ErrNotFound(err error) render.Renderer {
	eMsg, details := getErrorMessage(err)
	errorCode := getErrorCode(eMsg)

	return &Error{
		HTTPStatusCode: http.StatusNotFound,
		ErrorCode:      errorCode,
		ErrorText:      details,
	}
}

// Internal error. Deliberately returns a fixed, generic message to the client
// rather than the underlying error text: a 500 wraps driver/database errors
// whose raw messages can leak schema or internal details. The specific cause is
// still available server-side (handlers log the raw error before rendering).
func ErrInternal(err error) render.Renderer {
	eMsg, _ := getErrorMessage(err)
	errorCode := getErrorCode(eMsg)

	return &Error{
		HTTPStatusCode: http.StatusInternalServerError,
		ErrorCode:      errorCode,
		ErrorText:      "Internal server error",
	}
}

// Unauthorized error
func ErrUnauthorized(err error) render.Renderer {
	eMsg, details := getErrorMessage(err)
	errorCode := getErrorCode(eMsg)

	return &Error{
		HTTPStatusCode: http.StatusUnauthorized,
		ErrorCode:      errorCode,
		ErrorText:      details,
	}
}

// Conflict error (HTTP 409), e.g. a duplicate-key violation from trying to
// create a row that already exists.
func ErrConflict(err error) render.Renderer {
	eMsg, details := getErrorMessage(err)
	errorCode := getErrorCode(eMsg)

	return &Error{
		HTTPStatusCode: http.StatusConflict,
		ErrorCode:      errorCode,
		ErrorText:      details,
	}
}

// ErrFromRepo maps a repository-layer error to the HTTP response whose status
// best reflects its cause, so that a not-found row surfaces as 404 (not 500), a
// duplicate as 409, and a validation/foreign-key problem as 400. Errors that
// aren't a recognized repository error (template failures, tx.Commit, etc.)
// fall through to 500. Handlers should render repository errors through this
// instead of hard-coding ErrInternal, which previously turned every bad id into
// a 500.
func ErrFromRepo(err error) render.Renderer {
	switch {
	case strings.Contains(err.Error(), string(NotFoundError)):
		return ErrNotFound(err)
	case strings.Contains(err.Error(), string(DuplicateKeyError)):
		return ErrConflict(err)
	case strings.Contains(err.Error(), string(ValidationError)),
		strings.Contains(err.Error(), string(ForeignKeyError)):
		return ErrBadRequest(err)
	default:
		return ErrInternal(err)
	}
}
