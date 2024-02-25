package repositories

import (
	"context"

	"github.com/bataloff/tiknkoff/internal/models"
	builder "github.com/doug-martin/goqu/v9"

	"github.com/bataloff/tiknkoff/domain"
	"github.com/bataloff/tiknkoff/pkg/database"
)

type OrderBookRepository struct {
	pool database.Pool
}

func NewOrderBookRepository(ctx context.Context, pool database.Pool) (*OrderBookRepository, error) {
	o := &OrderBookRepository{
		pool: pool,
	}
	if err := o.migrate(ctx); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *OrderBookRepository) migrate(ctx context.Context) error {
	const schema = `
create table if not exists orderbooks (
    id integer primary key autoincrement,
    figi text,
    instrument_uid text,
    depth integer,
    is_consistent integer,
    time_unix integer,
    limit_up real,
    limit_down real
);

create table if not exists bids (
    orderbook_id integer,
    price real,
    quantity integer,
    foreign key (orderbook_id) references orderbooks (id) on delete cascade 
);

create table if not exists asks (
    orderbook_id integer,
    price real,
    quantity integer,
    foreign key (orderbook_id) references orderbooks (id) on delete cascade 
);
`
	if _, err := o.pool.Builder().ExecContext(ctx, schema); err != nil {
		return err
	}
	return nil
}

func (o *OrderBookRepository) Store(ctx context.Context, orders []*domain.OrderBook) error {
	return o.pool.Builder().WithTx(func(tx *builder.TxDatabase) error {
		for i := range orders {
			order := models.OrderBookFromEntity(orders[i])

			result, err := tx.Insert("orderbooks").Rows(order).Executor().ExecContext(ctx)
			if err != nil {
				return err
			}

			lastID, err := result.LastInsertId()
			if err != nil {
				return err
			}

			bids, asks := models.OrdersFromEntity(lastID, orders[i])

			_, err = tx.Insert("bids").Rows(bids).Executor().ExecContext(ctx)
			if err != nil {
				return err
			}

			_, err = tx.Insert("asks").Rows(asks).Executor().ExecContext(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
