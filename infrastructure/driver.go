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

const maxOpenDBConn = 10
const maxIdleDBConn = 5
const maxDBConnLifetime = 5 * time.Minute

// NewDatabase 新しいPostgresDBの接続プールを作成
func NewDatabase(dsn string) *DB {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	if err = conn.Ping(); err != nil {
		panic(err)
	}

	conn.SetMaxOpenConns(maxOpenDBConn)
	conn.SetMaxIdleConns(maxIdleDBConn)
	conn.SetConnMaxLifetime(maxDBConnLifetime)

	return &DB{conn}
}
