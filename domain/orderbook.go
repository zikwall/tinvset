package domain

type Order struct {
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
}

type OrderBook struct {
	Figi          string  `json:"Figi"`
	InstrumentUid string  `json:"InstrumentUid"`
	Depth         int32   `json:"Depth"`
	IsConsistent  bool    `json:"IsConsistent"`
	TimeUnix      int64   `json:"TimeUnix"`
	LimitUp       float64 `json:"LimitUp"`
	LimitDown     float64 `json:"LimitDown"`
	Bids          []Order `json:"Bids"`
	Asks          []Order `json:"Asks"`
}
