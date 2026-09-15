package xsql

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// bufLogger is a minimal Logger that records everything it receives.
type bufLogger struct {
	buf bytes.Buffer
}

func (l *bufLogger) Printf(format string, v ...any) {
	l.buf.WriteString(strings.TrimSpace(fmt.Sprintf(format, v...)))
	l.buf.WriteString("\n")
}

// fakeDBI is a minimal DBI whose only purpose is to let Debug log calls.
type fakeDBI struct{}

func (fakeDBI) ExecContext(context.Context, string, ...any) (sql.Result, error) { return nil, nil }
func (fakeDBI) QueryContext(context.Context, string, ...any) (*sql.Rows, error) { return nil, nil }
func (fakeDBI) Begin() (*sql.Tx, error)                                         { return nil, nil }
func (fakeDBI) BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)        { return nil, nil }

// failDBI always fails, so Debug's error logging can be verified.
type failDBI struct{}

func (failDBI) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, errors.New("exec boom")
}
func (failDBI) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errors.New("query boom")
}
func (failDBI) Begin() (*sql.Tx, error)                                  { return nil, nil }
func (failDBI) BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error) { return nil, nil }

func TestDebugWithLogger(t *testing.T) {
	var l bufLogger
	db := Debug(fakeDBI{}, WithLogger(&l))

	_, _ = db.ExecContext(context.Background(), "INSERT INTO t VALUES (?)", 1)
	_, _ = db.QueryContext(context.Background(), "SELECT * FROM t")

	got := l.buf.String()
	if !strings.Contains(got, "Debug Exec") || !strings.Contains(got, "query:INSERT INTO t VALUES (?)") {
		t.Errorf("exec not logged: %q", got)
	}
	if !strings.Contains(got, "Debug Query") || !strings.Contains(got, "query:SELECT * FROM t") {
		t.Errorf("query not logged: %q", got)
	}
}

func TestDebugLogsError(t *testing.T) {
	var l bufLogger
	db := Debug(failDBI{}, WithLogger(&l))

	if _, err := db.ExecContext(context.Background(), "INSERT INTO t VALUES (?)", 1); err == nil {
		t.Fatal("expected exec error")
	}
	if _, err := db.QueryContext(context.Background(), "SELECT * FROM t"); err == nil {
		t.Fatal("expected query error")
	}

	got := l.buf.String()
	if !strings.Contains(got, "Debug Exec error") || !strings.Contains(got, "exec boom") {
		t.Errorf("exec error not logged: %q", got)
	}
	if !strings.Contains(got, "Debug Query error") || !strings.Contains(got, "query boom") {
		t.Errorf("query error not logged: %q", got)
	}
}

func TestSlogLogger(t *testing.T) {
	var buf bytes.Buffer
	l := SlogLogger(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	l.Printf("hello %s", "world")
	if got := buf.String(); !strings.Contains(got, "hello world") {
		t.Errorf("slog adapter did not log expected output: %q", got)
	}
}

func TestDBSetLogger(t *testing.T) {
	var l bufLogger
	db, err := NewDB("sqlite", &Config{DSN: "file::memory:?cache=shared", Active: 1, Idle: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Master().Close()

	db.SetLogger(&l)
	if db.Logger() == nil {
		t.Fatal("expected logger to be configured")
	}

	if _, err := db.ExecContext(context.Background(), "CREATE TABLE t (id INTEGER)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.QueryContext(context.Background(), "SELECT * FROM t"); err != nil {
		t.Fatal(err)
	}

	got := l.buf.String()
	if !strings.Contains(got, "Exec") || !strings.Contains(got, "CREATE TABLE t (id INTEGER)") {
		t.Errorf("exec not logged: %q", got)
	}
	if !strings.Contains(got, "Query") || !strings.Contains(got, "SELECT * FROM t") {
		t.Errorf("query not logged: %q", got)
	}

	// A failing statement must also be logged.
	if _, err := db.QueryContext(context.Background(), "SELECT * FROM missing_table"); err == nil {
		t.Fatal("expected query error")
	}
	if got := l.buf.String(); !strings.Contains(got, "Query error") ||
		!strings.Contains(got, "SELECT * FROM missing_table") {
		t.Errorf("query error not logged: %q", got)
	}

	// Runtime reconfiguration, including disabling logging.
	db.SetLogger(nil)
	if db.Logger() != nil {
		t.Fatal("expected logger to be disabled")
	}
}
