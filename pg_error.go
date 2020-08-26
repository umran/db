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
	// We look for either:
	//  - the standard PG errcode SerializationFailureError:40001 or
	//  - the Cockroach extension errcode RetriableError:CR000. This extension
	//    has been removed server-side, but support for it has been left here for
	//    now to maintain backwards compatibility.
	code := errCode(err)

	// debugging
	if !(code == "CR000" || code == "40001") {
		// fmt.Println(errorCause(err))
	}

	return code == "CR000" || code == "40001"
}

func errCode(err error) string {
	switch t := errorCause(err).(type) {
	case *pq.Error:
		return string(t.Code)

	case errWithSQLState:
		return t.SQLState()

	default:
		return ""
	}
}

// errWithSQLState is implemented by pgx (pgconn.PgError).
//
// TODO(andrei): Add this method to pq.Error and stop depending on lib/pq.
type errWithSQLState interface {
	SQLState() string
}
