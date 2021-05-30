package infrastructure

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB struct {
	*sql.DB
}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// NewDatabase creates a new database pool for the application
func NewDatabase(dsn string) *DB {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	if err = conn.Ping(); err != nil {
		panic(err)
	}

	conn.SetMaxOpenConns(maxOpenDbConn)
	conn.SetMaxIdleConns(maxIdleDbConn)
	conn.SetConnMaxLifetime(maxDbLifetime)

	return &DB{conn}
}
