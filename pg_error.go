package db

import (
	"github.com/lib/pq"
)

// errorCause returns the original cause of the error, if possible. An
// error has a proximate cause if it's type is compatible with Go's
// errors.Unwrap() or pkg/errors' Cause(); the original cause is the
// end of the causal chain.
func errorCause(err error) error {
	for err != nil {
		if c, ok := err.(interface{ Cause() error }); ok {
			err = c.Cause()
		} else if c, ok := err.(interface{ Unwrap() error }); ok {
			err = c.Unwrap()
		} else {
			break
		}
	}

	return err
}

func errIsRetryable(err error) bool {
	code := errCode(err)

	return code == "40001"
}

func errCode(err error) string {
	switch t := errorCause(err).(type) {
	case *pq.Error:
		return string(t.Code)
	default:
		return ""
	}
}
