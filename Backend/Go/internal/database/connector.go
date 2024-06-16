package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
)

type SQLQueryExec interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SQLTXQueryExec interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func OpenDb(ctx context.Context, connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Создание таблиц, если они не существуют
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username varchar(24) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			last_ram_generated INT,
			avatar_ram_id INT REFERENCES rams (id),
			avatar_box BOX 
		);
		CREATE TABLE IF NOT EXISTS rams (
		    id SERIAL PRIMARY KEY,
		    name TEXT DEFAULT '',
		    image_url TEXT NOT NULL,
		    user_id INT NOT NULL REFERENCES users (id)
		)
	`)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
