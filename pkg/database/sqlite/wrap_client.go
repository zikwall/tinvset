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

type ConnectionPool struct {
	db *builder.Database
}

func (c *ConnectionPool) Builder() *builder.Database {
	return c.db
}

// Drop close not implemented in database
func (c *ConnectionPool) Drop() error {
	return nil
}

func (c *ConnectionPool) DropMsg() string {
	return "close database: this is not implemented"
}

func New(ctx context.Context, path string, debug bool) (*ConnectionPool, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = db.PingContext(pingCtx); err != nil {
		return nil, err
	}

	dialect := builder.Dialect("sqlite3")
	connect := &ConnectionPool{
		db: dialect.DB(db),
	}

	if debug {
		logger := &database.Logger{}
		logger.SetCallback(func(format string, v ...interface{}) {
			log.Panicln(v)
		})
		connect.db.Logger(logger)
	}

	return connect, nil
}
