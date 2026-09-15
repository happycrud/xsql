package xsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
)

// Logger records executed SQL. It is satisfied by *log.Logger and most loggers.
type Logger interface {
	Printf(format string, v ...any)
}

// DebugOption configures the DebugDB returned by Debug.
type DebugOption func(*DebugDB)

// WithLogger overrides the default logger (log.Default()); nil keeps it.
func WithLogger(l Logger) DebugOption {
	return func(d *DebugDB) {
		if l != nil {
			d.log = l
		}
	}
}

// Debug wraps db and logs every executed SQL statement and error. It defaults
// to log.Default(); use WithLogger to route output elsewhere.
func Debug(db DBI, opts ...DebugOption) *DebugDB {
	d := &DebugDB{dbt: db, log: log.Default()}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// SlogLogger adapts a *slog.Logger to Logger using the debug level.
func SlogLogger(l *slog.Logger) Logger {
	if l == nil {
		l = slog.Default()
	}
	return slogLogger{l}
}

type slogLogger struct{ l *slog.Logger }

func (s slogLogger) Printf(format string, v ...any) {
	s.l.Debug(fmt.Sprintf(format, v...))
}

type DebugDB struct {
	log Logger
	dbt DBI
}

type DebugTx struct {
	log Logger
	eq  ExecQuerier
}

func (d *DebugDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	d.log.Printf("Debug Exec ctx:%v query:%s args:%+v", ctx, query, args)
	r, err := d.dbt.ExecContext(ctx, query, args...)
	if err != nil {
		d.log.Printf("Debug Exec error ctx:%v query:%s args:%+v err:%v", ctx, query, args, err)
	}
	return r, err
}

func (d *DebugDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	d.log.Printf("Debug Query ctx:%v query:%s args:%+v", ctx, query, args)
	r, err := d.dbt.QueryContext(ctx, query, args...)
	if err != nil {
		d.log.Printf("Debug Query error ctx:%v query:%s args:%+v err:%v", ctx, query, args, err)
	}
	return r, err
}

func (d *DebugDB) Begin() (*DebugTx, error) {
	tx, err := d.dbt.Begin()
	if err != nil {
		return nil, err
	}
	return &DebugTx{eq: tx, log: d.log}, nil
}

func (d *DebugDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*DebugTx, error) {
	tx, err := d.dbt.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &DebugTx{eq: tx, log: d.log}, nil
}

func (d *DebugTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	d.log.Printf("Debug Tx Exec ctx:%v query:%s args:%+v", ctx, query, args)
	r, err := d.eq.ExecContext(ctx, query, args...)
	if err != nil {
		d.log.Printf("Debug Tx Exec error ctx:%v query:%s args:%+v err:%v", ctx, query, args, err)
	}
	return r, err
}

func (d *DebugTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	d.log.Printf("Debug TX Query ctx:%v query:%s args:%+v", ctx, query, args)
	r, err := d.eq.QueryContext(ctx, query, args...)
	if err != nil {
		d.log.Printf("Debug TX Query error ctx:%v query:%s args:%+v err:%v", ctx, query, args, err)
	}
	return r, err
}
