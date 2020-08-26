package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cockroachdb/cockroach-go/crdb"
)

// CRDBConnection ...
type CRDBConnection struct {
	db *sql.DB
}

// Query ...
func (cr *CRDBConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return cr.db.Query(query, args...)
}

// QueryRow ...
func (cr *CRDBConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return cr.db.QueryRow(query, args...)
}

// Exec ...
func (cr *CRDBConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return cr.db.Exec(query, args...)
}

// ExecTx ...
func (cr *CRDBConnection) ExecTx(handler func(Transaction) error) error {
	return crdb.ExecuteTx(context.Background(), cr.db, nil, func(tx *sql.Tx) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = errors.New("panicked during transaction")
			}
		}()

		err = handler(tx)
		return
	})
}

// NewCRDBConnection ...
func NewCRDBConnection(uri string) (*CRDBConnection, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	return &CRDBConnection{
		db: db,
	}, nil
}
