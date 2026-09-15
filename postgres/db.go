package postgres

import (
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/goflower-io/xsql"
)

func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("pgx", c)
}

var PgxMap = sync.OnceValue(func() *pgtype.Map {
	return pgtype.NewMap()
})
