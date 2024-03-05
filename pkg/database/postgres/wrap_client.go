package postgres

import (
	"context"
	"log"
	"time"

	// nolint:revive // it's OK
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/lib/pq"

	builder "github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"

	"github.com/bataloff/tiknkoff/pkg/database"
)

type Pool struct {
	db *builder.Database
}

func (c *Pool) Dialect() string {
	return "postgres"
}

func (c *Pool) Builder() *builder.Database {
	return c.db
}

// Drop close not implemented in database
func (c *Pool) Drop() error {
	return nil
}

func (c *Pool) DropMsg() string {
	return "close database: is not implemented"
}
func NewPool(ctx context.Context, opt *database.Opt) (*Pool, error) {
	db, err := sqlx.Open(opt.Dialect, opt.ConnectionString())
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = db.PingContext(pingCtx); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(opt.MaxIdleConns)
	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetConnMaxLifetime(opt.MaxConnMaxLifetime)

	dialect := builder.Dialect(opt.Dialect)
	pool := dialect.DB(db)

	if opt.Debug {
		logger := &database.Logger{}
		logger.SetCallback(func(format string, v ...interface{}) {
			log.Println(v)
		})
		pool.Logger(logger)
	}

	return &Pool{db: pool}, nil
}
