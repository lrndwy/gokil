package orm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

type Tx struct {
	*sql.Tx
	db *DB
}

type dbContextKey struct{}

func Connect(driver, dsn string, maxOpen, maxIdle int) (*DB, error) {
	switch driver {
	case "postgres":
		driver = "pgx"
	case "mysql":
		driver = "mysql"
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &DB{DB: db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func WithDB(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, dbContextKey{}, db)
}

func DBFromContext(ctx context.Context) *DB {
	if db, ok := ctx.Value(dbContextKey{}).(*DB); ok {
		return db
	}
	return nil
}

func WithTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error {
	db := DBFromContext(ctx)
	if db == nil {
		return fmt.Errorf("no database in context")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	wrapped := &Tx{Tx: tx, db: db}
	txCtx := context.WithValue(ctx, dbContextKey{}, &DB{DB: db.DB})

	if err := fn(txCtx, wrapped); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func connFromContext(ctx context.Context) queryConn {
	db := DBFromContext(ctx)
	if db == nil {
		return nil
	}
	if tx, ok := ctx.Value(txContextKey{}).(*Tx); ok {
		return tx
	}
	return db
}

type queryConn interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type txContextKey struct{}

func withTxContext(ctx context.Context, tx *Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}
