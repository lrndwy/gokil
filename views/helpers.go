package views

import (
	"database/sql"
	"errors"
)

// NotFoundIf maps sql.ErrNoRows to a 404 HTTPError. Other errors are returned unchanged.
func NotFoundIf(err error, notFoundMessage string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound(notFoundMessage)
	}
	return err
}
