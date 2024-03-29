package sqlite

import (
	"context"
	"database/sql"
	"log"
	"time"

	builder "github.com/doug-martin/goqu/v9"

	// nolint:golint // it's OK
	// _ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/go-sql-driver/mysql"

	"github.com/bataloff/tiknkoff/pkg/database"
)

const sqlite3Dialect = "sqlite3"

type Pool struct {
	db *builder.Database
}

func (c *Pool) Dialect() string {
	return sqlite3Dialect
}

func (c *Pool) Builder() *builder.Database {
	return c.db
}

// Drop close not implemented in database
func (c *Pool) Drop() error {
	return nil
}

func (c *Pool) DropMsg() string {
	return "close database: this is not implemented"
}

func NewPool(ctx context.Context, opt *database.Opt) (*Pool, error) {

	db, err := sql.Open(opt.Dialect, opt.SQLitePath)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = db.PingContext(pingCtx); err != nil {
		return nil, err
	}

	dialect := builder.Dialect(opt.Dialect)
	connect := &Pool{
		db: dialect.DB(db),
	}

	if opt.Debug {
		logger := &database.Logger{}
		logger.SetCallback(func(format string, v ...interface{}) {
			log.Panicln(v)
		})
		connect.db.Logger(logger)
	}

	return connect, nil
}
