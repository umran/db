# DB - A Library for Executing SQL Transactions in the Serializable Isolation Level with Automatic Retry Handling

This is a little helper library I wrote to extend a standard Go SQL connection to more robustly support transactions under the serializable isolation level. By robust support I mean transactions that automatically retry on retriable commit errors, which sometimes occur when multiple transactions attempt to commit at the same time under the serializable isolation level. Since these errors are transient and are usually resolved upon retry, automatic retry behaviour is a useful feature for some applications that would rather the database driver handle such errors.

Currently this library only supports PostgreSQL and CockroachDB.

# Usage
## Creating a Connection
We start by creating a connection to a database.

To connect to PostgreSQL:
```go
conn, err := NewPGConnection("connection_uri")
```

To connect to CockroachDB:
```go
conn, err := NewCRDBConnection("connection_uri")
```

The returned connection instance supports the following methods exactly as `*sql.DB` would:
- `Query(query string, args ...interface{}) (*sql.Rows, error)`
- `QueryRow(query string, args ...interface{}) *sql.Row`
- `Exec(query string, args ...interface{}) (sql.Result, error)`

## Executing a Transaction in the Serializable Isolation Level
In order to execute a transaction in the serializable isolation level we must call a special method called `ExecTx` which has the following signature:
- `ExecTx(func(tx db.Transaction) error) error`

Notice how its argument is a function with the signature, `func(tx db.Transaction) error`.
This function receives an argument of type `db.Transaction` and returns an error.

`db.Transaction` is an interface that implements the following methods exactly as `*sql.DB` would:
- `Query(query string, args ...interface{}) (*sql.Rows, error)`
- `QueryRow(query string, args ...interface{}) *sql.Row`
- `Exec(query string, args ...interface{}) (sql.Result, error)`

All operations that should be part of the same transaction must happen inside this function.
If an error is returned, the entire transaction is aborted and rolled back.

An example transaction:
```go
err := conn.ExecTx(func(tx db.Transaction) error {
    var age int
    if err := tx.QueryRow("SELECT age FROM person WHERE last_name = $1", "Jones").Scan(&age); err != nil {
        fmt.Println("the query was not sucessful, aborting transaction...")
        return err
    }

    fmt.Println("the query was successful, attempting to commit transaction...")
    return nil
})

if err != nil {
    fmt.Println("the transaction was aborted due to an error")
}

fmt.Println("the transaction was committed successfully")
```