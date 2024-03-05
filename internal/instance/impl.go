package instance

import (
	"context"

	"github.com/tinkoff/invest-api-go-sdk/investgo"

	"github.com/bataloff/tiknkoff/pkg/database"
	"github.com/bataloff/tiknkoff/pkg/database/postgres"
	"github.com/bataloff/tiknkoff/pkg/database/sqlite"
	"github.com/bataloff/tiknkoff/pkg/drop"
	"github.com/bataloff/tiknkoff/pkg/invest"
	"github.com/bataloff/tiknkoff/pkg/logger"
)

type Instance struct {
	*drop.Impl

	Pool database.Pool

	Logger *logger.Logger
	Client *invest.Client
}

func New(ctx context.Context, opt *Options) (*Instance, error) {
	s := &Instance{}
	s.Impl = drop.NewContext(ctx)

	var err error

	switch opt.Database.Dialect {
	case "postgres":
		s.Pool, err = postgres.NewPool(s.Context(), opt.Database)
		if err != nil {
			return nil, err
		}
		s.AddDropper(s.Pool.(*postgres.Pool))
	case "sqlite3":
		s.Pool, err = sqlite.NewPool(s.Context(), opt.Database)
		if err != nil {
			return nil, err
		}
		s.AddDropper(s.Pool.(*sqlite.Pool))
	}

	s.Logger, err = logger.New()
	if err != nil {
		return nil, err
	}
	s.AddDropper(s.Logger)

	client, err := investgo.NewClient(s.Context(), opt.Invest, s.Logger.Sugar())
	if err != nil {
		return nil, err
	}

	s.Client = invest.New(client)
	s.AddDropper(s.Client)

	return s, nil
}
