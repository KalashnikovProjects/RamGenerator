package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KalashnikovProjects/RamGenerator/Backend/Go-Api/internal/config"
	_ "github.com/lib/pq"
	"log"
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

// GenerateQueryAndArgsForUpdate генерирует строку запроса и аргументы для вставки
// table - название таблицы
// fields - хэш-мапа в формате поля в бд и значения для текущего запроса (если nil или default value оно игнорируется), пример:
//
//	 map[string]any{
//			"username":           user.Username,
//			"password_hash":      user.PasswordHash}
//
// condition - условие типа id=$1
// conditionValues - список значений для условия в порядке из условия.
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
		fmt.Println(err)
		return nil, err
	}

	// Создание таблиц, если они не существуют
	_, err = db.ExecContext(ctx, `
		CREATE EXTENSION IF NOT EXISTS CITEXT;
		CREATE TABLE IF NOT EXISTS users (
			id 						  SERIAL PRIMARY KEY,
			username 				  CITEXT UNIQUE NOT NULL,
			password_hash 			  TEXT NOT NULL,
			daily_ram_generation_time INT  NOT NULL DEFAULT 0,
			rams_generated_last_day   INT  NOT NULL DEFAULT 0,
			avatar_ram_id 			  INT NOT NULL,
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
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func GeneratePostgresConnectionString(user, password, host string, pgPort int, dbName string) string {
	return fmt.Sprintf(`postgresql://%s:%s@%s:%d/%s?sslmode=disable`, user, password, host, pgPort, dbName)
}

func CreateDBConnectionContext(ctx context.Context) *sql.DB {
	var db *sql.DB
	var err error
	connectionString := GeneratePostgresConnectionString(config.Conf.Database.User, config.Conf.Database.Password, config.Conf.Database.Hostname, config.Conf.Database.Port, config.Conf.Database.DBName)
	for {
		db, err = OpenDb(ctx, connectionString)
		if err == nil {
			break
		}
		log.Print("retry db connection, error: ", err)
		time.Sleep(2 * time.Second)
	}
	log.Print("successful connected to db")
	return db
}
