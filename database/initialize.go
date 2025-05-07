package database

import (
	"context"
	"database/sql"
	"fmt"
)

func CreateTable(ctx context.Context, port, dbName, password, tableName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:%s)/%s?parseTime=true", dbName, password, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Wait for the DB to be ready (optional backoff can be added)
	if err := db.PingContext(ctx); err != nil {
		return err
	}

	createTableExec := fmt.Sprintf(`CREATE TABLE %s (
			id INT AUTO_INCREMENT NOT NULL,
			term VARCHAR(128) NOT NULL,
			definition VARCHAR(255) NOT NULL,
			PRIMARY KEY (id)
		)`, tableName)
	_, err = db.ExecContext(ctx, createTableExec)

	return err
}
