package repositories

import "errors"

// ErrEntityNotFound is returned when a particular entity cannot be found in the database
var ErrEntityNotFound = errors.New("entity not found")
