package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/krtffl/torro/internal/domain"
)

func getTableName(err error) string {
	return strings.Split(err.Error(), "\"")[1]
}

func getColumnAndTableNames(err error) (string, string) {
	column := strings.Split(err.Error(), "\"")[1]
	table := strings.Split(err.Error(), "\"")[3]
	return column, table
}

func handleErrors(err error) error {
	// Handle nil error
	if err == nil {
		return nil
	}

	// Classify PostgreSQL errors by their SQLSTATE code (via lib/pq's
	// structured *pq.Error) rather than by pattern-matching the rendered
	// message. lib/pq's Error() only ever returns "pq: <Message>" - it
	// never includes an "ERROR: ..." prefix or a "(SQLSTATE ...)" suffix -
	// so the previous message-based checks here never actually matched a
	// real driver error and silently fell through to errUnknown. See
	// https://pkg.go.dev/github.com/lib/pq#Error and the PostgreSQL error
	// codes appendix for the codes below.
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "42P01": // undefined_table
			return errTableNotExistent(getTableName(err))
		case "42703": // undefined_column
			return errColumnNotExistent(getColumnAndTableNames(err))
		case "23505": // unique_violation
			return errDuplicateKey()
		case "23503": // foreign_key_violation
			return errForeignKeyConstraint()
		}
	}

	if strings.Contains(err.Error(), "record not found") {
		return errNotFound()
	}

	if strings.Contains(err.Error(), "no rows in result set") {
		return errNotFound()
	}

	return errUnknown(err)
}

func errTableNotExistent(table string) error {
	return fmt.Errorf("%s: Table %s does not exist", domain.NonExistentTableError, table)
}

func errColumnNotExistent(column, table string) error {
	return fmt.Errorf(
		"%s: Column %s does not exist on table %s",
		domain.NonExistentColumnError,
		column,
		table,
	)
}

func errDuplicateKey() error {
	return fmt.Errorf("%s: Duplicate key", domain.DuplicateKeyError)
}

func errForeignKeyConstraint() error {
	return fmt.Errorf("%s: Foreign key violation", domain.ForeignKeyError)
}

func errNotFound() error {
	return fmt.Errorf("%s: Record not found", domain.NotFoundError)
}

func errUnknown(err error) error {
	return fmt.Errorf("%s: %v", domain.UnknownError, err)
}
