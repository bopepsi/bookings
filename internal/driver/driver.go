package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Connect app to databse

// DB holds databse connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const (
	maxOpenDbConn = 10
	maxIdleDbConn = 5
	maxDbLifetime = 5 * time.Minute
)

//dsn: string to connect to databse
// Create database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	db, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(maxOpenDbConn)
	db.SetMaxIdleConns(maxIdleDbConn)
	db.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = db

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// Create new databse for our app
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
