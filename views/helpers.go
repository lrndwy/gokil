package views

import (
	"database/sql"
	"errors"
)

func NotFoundIf(err error, notFoundMessage string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return &HTTPError{Status: 404, Message: notFoundMessage}
	}
	return err
}
