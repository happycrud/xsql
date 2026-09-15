// Package sqlite provides the default SQLite driver for xsql.
//
// It is backed by github.com/ncruces/go-sqlite3, a pure-Go (WebAssembly)
// SQLite implementation that requires no cgo.
//
// Alternative drivers are available as sub-packages:
//
//	github.com/goflower-io/xsql/sqlite/mattn   // github.com/mattn/go-sqlite3 (cgo)
//	github.com/goflower-io/xsql/sqlite/modernc // modernc.org/sqlite (pure Go)
package sqlite

import (
	_ "github.com/ncruces/go-sqlite3/driver"

	"github.com/goflower-io/xsql"
)

// NewDB opens a SQLite database using the default (ncruces) driver.
func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("sqlite3", c)
}
