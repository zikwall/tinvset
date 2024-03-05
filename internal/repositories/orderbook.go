package repositories

import (
	"context"
	"fmt"

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
	// все это нужно вынести потом в миграции
	if _, err := o.pool.Builder().ExecContext(ctx, getSchema(o.pool.Dialect())); err != nil {
		return err
	}
	return nil
}

func (o *OrderBookRepository) Store(ctx context.Context, orders []*domain.OrderBook) error {
	return o.pool.Builder().WithTx(func(tx *builder.TxDatabase) error {
		for i := range orders {
			order := models.OrderBookFromEntity(orders[i])

			var lastInsertID int64

			switch o.pool.Dialect() {
			case "postgres":
				insert := tx.Insert("orderbooks").Rows(order).Returning("id").Executor()
				if _, err := insert.ScanVal(&lastInsertID); err != nil {
					return err
				}
			case "sqlite3":
				result, err := tx.Insert("orderbooks").Rows(order).Executor().ExecContext(ctx)
				if err != nil {
					return err
				}

				lastInsertID, err = result.LastInsertId()
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported insert operation for dialect %s", o.pool.Dialect())
			}

			bids, asks := models.OrdersFromEntity(lastInsertID, orders[i])
			if _, err := tx.Insert("bids").Rows(bids).Executor().ExecContext(ctx); err != nil {
				return err
			}

			if _, err := tx.Insert("asks").Rows(asks).Executor().ExecContext(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
