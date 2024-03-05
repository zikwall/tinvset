package models

import "github.com/bataloff/tiknkoff/domain"

// В целом тут можно не конвертировать, но видим, что структура базы отличается от сущности домена
// поэтому оставим возможность конвертирования из домена в модели.

type Order struct {
	OrderBookID int64   `db:"orderbook_id"`
	Price       float64 `db:"price"`
	Quantity    int64   `db:"quantity"`
}

func OrdersFromEntity(obID int64, order *domain.OrderBook) ([]Order, []Order) {
	bids := toOrderModels(obID, order.Bids)
	asks := toOrderModels(obID, order.Asks)
	return bids, asks
}

func toOrderModels(obID int64, orders []domain.Order) []Order {
	mod := make([]Order, len(orders))
	for i := range orders {
		mod[i] = Order{
			OrderBookID: obID,
			Price:       orders[i].Price,
			Quantity:    orders[i].Quantity,
		}
	}
	return mod
}

type OrderBook struct {
	Figi          string  `db:"figi"`
	InstrumentUid string  `db:"instrument_uid"`
	Depth         int32   `db:"depth"`
	IsConsistent  uint8   `db:"is_consistent"`
	TimeUnix      int64   `db:"time_unix"`
	LimitUp       float64 `db:"limit_up"`
	LimitDown     float64 `db:"limit_down"`
}

func OrderBookFromEntity(order *domain.OrderBook) OrderBook {
	return OrderBook{
		Figi:          order.Figi,
		InstrumentUid: order.InstrumentUid,
		Depth:         order.Depth,
		IsConsistent:  b2u8(order.IsConsistent),
		TimeUnix:      order.TimeUnix,
		LimitUp:       order.LimitUp,
		LimitDown:     order.LimitDown,
	}
}

func b2u8(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}
