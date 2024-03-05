package database

import builder "github.com/doug-martin/goqu/v9"

// Pool is common database pool interface
type Pool interface {
	Builder() *builder.Database
	Dialect() string
}
