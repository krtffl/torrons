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
	// PostgreSQL errors
	NonExistentTableError  ErrorMsg = "NonExistentPostgreSQLTableError"
	NonExistentColumnError ErrorMsg = "NonExistentPostgreSQLColumnError"
	DuplicateKeyError      ErrorMsg = "DuplicateKeyPostgreSQLError"
	ForeignKeyError        ErrorMsg = "ForeignKeyPostgreSQLError"
	NotFoundError          ErrorMsg = "NotFoundPostgreSQLError"
	UnknownError           ErrorMsg = "UnkwnownPostgreSQLError"
)

var ErrorCodes = map[ErrorMsg]ErrorCode{
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

// Internal error
func ErrInternal(err error) render.Renderer {
	eMsg, details := getErrorMessage(err)
	errorCode := getErrorCode(eMsg)

	return &Error{
		HTTPStatusCode: http.StatusInternalServerError,
		ErrorCode:      errorCode,
		ErrorText:      details,
	}
}
