package db

import (
	"database/sql"
)

// PGConnection ...
type PGConnection struct {
	db *sql.DB
}

// Query ...
func (pg *PGConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return pg.db.Query(query, args...)
}

// QueryRow ...
func (pg *PGConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return pg.db.QueryRow(query, args...)
}

// Exec ...
func (pg *PGConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return pg.db.Exec(query, args...)
}

// ExecTx ...
func (pg *PGConnection) ExecTx(handler func(Transaction) error) (err error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	for {
		_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
		if err != nil {
			return
		}

		err = handler(tx)
		if err == nil {
			// attempt to commit, keeping in mind it might result in a retryable error
			// return nil if and only if there is no error, at which point the transaction has been committed
			if err = tx.Commit(); err == nil {
				return
			}
		}

		// We got an error; let's see if it's a retryable one and, if so, restart.
		if !errIsRetryable(err) {
			return
		}

		// the error is retryable, so abort the current transaction and start a new one
		tx.Rollback()

		var newTx *sql.Tx
		newTx, err = pg.db.Begin()
		if err != nil {
			return
		}

		tx = newTx
	}
}

// NewPGConnection ...
func NewPGConnection(uri string) (*PGConnection, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	return &PGConnection{
		db: db,
	}, nil
}
