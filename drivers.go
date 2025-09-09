package main

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	duckdb "github.com/marcboeker/go-duckdb/v2"
)

////////// INITIALIZE DATABASE //////////

func NewDuckSack(dbpath string, initsql string, vector_dimensions int, related_eps float64) *BeanSack {
	// conn, err := duckdb.NewConnector(fmt.Sprintf("%s?threads=%d", dbpath, max(1, runtime.NumCPU()-1)), nil)
	conn, err := duckdb.NewConnector(dbpath, nil)
	noerror(err, "CONNECTOR ERROR")

	// open connection
	db := sql.OpenDB(conn)
	if initsql != "" {
		_, err = db.Exec(fmt.Sprintf(initsql, vector_dimensions, related_eps))
		noerror(err, "INIT SQL ERROR")
	}

	return &BeanSack{
		connector: conn,
		db:        db,
		query:     sqlx.NewDb(db, "duckdb"),
		dim:       vector_dimensions,
	}
}

func NewSqliteSack(dbpath string, initsql string, vector_dimensions int, related_eps float64) *BeanSack {
	db, err := sql.Open("sqlite3", dbpath)
	noerror(err, "SQLITE DB ERROR")
	if initsql != "" {
		_, err = db.Exec(fmt.Sprintf(initsql, vector_dimensions, related_eps))
		noerror(err, "INIT SQL ERROR")
	}

	return &BeanSack{
		connector: nil,
		db:        db,
		query:     sqlx.NewDb(db, "sqlite3"),
		dim:       vector_dimensions,
	}
}
