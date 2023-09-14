package driver

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

type DBCreds struct {
	Host     string
	Port     int
	Username string
	Password string
}

var dbConn = &DB{}
var clickhouseConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

// ConnectSQL creates database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = d

	err = testDB(d)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

func ConnectClickhouse(dsn string) (*DB, error) {
	d, err := NewClickhouseDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	clickhouseConn.SQL = d

	err = testDB(d)
	if err != nil {
		return nil, err
	}
	return clickhouseConn, nil

}

// testDB tries to ping the database
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	var version string
	err = d.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Fatal(err)
	}

	// Print the DB version
	fmt.Printf("DB Version: %s\n", version)
	return nil
}

// NewDatabase creates a new database for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// New Clickhouse DB for the app
func NewClickhouseDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
