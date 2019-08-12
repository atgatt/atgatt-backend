package errors

import "fmt"

type NotFoundError struct {
	err  string
	code int
}

func NotFound(errMsg string) error {
	return &NotFoundError{
		err:  errMsg,
		code: 404,
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.err)
}

func (e *NotFoundError) ResponseCode() int {
	return e.code
}
