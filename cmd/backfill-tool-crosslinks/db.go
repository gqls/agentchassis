package main

import (
	"database/sql"
	"fmt"
	"os"
)

// dbConn mirrors cmd/config-key-audit's connection helper deliberately: the
// same env vars, the same refusal to guess. An unset PG_CLIENTS_HOST returns
// (nil, nil) so the caller can say what it wants about a missing environment —
// a backfill that silently cannot connect is the failure this shape avoids.
func dbConn() (*sql.DB, error) {
	host := os.Getenv("PG_CLIENTS_HOST")
	if host == "" {
		return nil, nil
	}
	pw := os.Getenv("CLIENTS_DB_PASSWORD")
	if pw == "" {
		return nil, fmt.Errorf("PG_CLIENTS_HOST is set but CLIENTS_DB_PASSWORD is not — refusing to guess a connection")
	}
	port := os.Getenv("PG_CLIENTS_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("PG_CLIENTS_USER")
	if user == "" {
		user = "clients_user"
	}
	dbname := os.Getenv("PG_CLIENTS_DB")
	if dbname == "" {
		dbname = "clients_db"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pw, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}
