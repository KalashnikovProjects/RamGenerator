package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"reflect"
	"strings"
	"time"
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

// GenerateQueryAndArgsForUpdate generate query string and arguments for paste
// table - table name
// fields - hash-map "field name": field value:
//
//	 map[string]any{
//			"username":           user.Username,
//			"password_hash":      user.PasswordHash}
//
// condition - condition like id=$1
// conditionValues - values for conditions
func GenerateQueryAndArgsForUpdate(table string, fields map[string]any, condition string, conditionValues ...any) (string, []any) {
	var updates []string
	var args []any
	for key, val := range fields {
		if val != nil && !reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface()) {
			updates = append(updates, fmt.Sprintf("%s = $%d", key, len(conditionValues)+len(updates)+1))
			args = append(args, val)
		}
	}
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(updates, ", "), condition)
	return query, append(conditionValues, args...)
}

func OpenDb(ctx context.Context, connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		slog.Error("opening db error", slog.String("error", err.Error()))
		return nil, err
	}

	_, err = db.ExecContext(ctx, `
		CREATE EXTENSION IF NOT EXISTS CITEXT;
		CREATE TABLE IF NOT EXISTS users (
			id 						  SERIAL PRIMARY KEY,
			username 				  CITEXT UNIQUE NOT NULL,
			password_hash 			  TEXT NOT NULL,
			
			daily_ram_generation_time INT  NOT NULL DEFAULT 0,
			rams_generated_last_day   INT  NOT NULL DEFAULT 0,
			clickers_blocked_until    INT  NOT NULL DEFAULT 0,
			
			avatar_ram_id 			  INT NOT NULL DEFAULT 0,
			avatar_box 				  BOX  NOT NULL
		);
		CREATE TABLE IF NOT EXISTS rams (
		    id 		    SERIAL PRIMARY KEY,
		    taps 		INT  NOT NULL DEFAULT 0,
		    description TEXT NOT NULL DEFAULT '',
		    image_url   TEXT NOT NULL,
		    user_id 	INT  NOT NULL REFERENCES users (id)
		);
	`)
	if err != nil {
		slog.Error("db exec error", slog.String("error", err.Error()))
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		slog.Error("db ping error", slog.String("error", err.Error()))
		return nil, err
	}

	return db, nil
}

func GeneratePostgresConnectionString(user, password, host string, dbName string) string {
	return fmt.Sprintf(`postgresql://%s:%s@%s/%s?sslmode=disable`, user, password, host, dbName)
}

func CreateDBConnectionContext(ctx context.Context, connectionString string) *sql.DB {
	var db *sql.DB
	var err error
	slog.Info("connecting to db")

	for {
		db, err = OpenDb(ctx, connectionString)
		if err == nil {
			break
		}
		slog.Debug("retry db connection", slog.String("error", err.Error()))

		time.Sleep(2 * time.Second)
	}
	slog.Info("successful connected to db")
	return db
}
