// Package mattn provides the SQLite driver backed by
// github.com/mattn/go-sqlite3, which uses cgo.
//
// Prefer the default github.com/goflower-io/xsql/sqlite package unless cgo is
// required.
package mattn

import (
	"github.com/goflower-io/xsql"

	_ "github.com/mattn/go-sqlite3"
)

// NewDB opens a SQLite database using the mattn/go-sqlite3 driver.
func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("sqlite3", c)
}
