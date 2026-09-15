// Package modernc provides the SQLite driver backed by modernc.org/sqlite, a
// pure-Go (transpiled) SQLite implementation.
//
// Prefer the default github.com/goflower-io/xsql/sqlite package (ncruces)
// unless modernc is specifically required.
package modernc

import (
	_ "modernc.org/sqlite"

	"github.com/goflower-io/xsql"
)

// NewDB opens a SQLite database using the modernc.org/sqlite driver.
func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("sqlite", c)
}
