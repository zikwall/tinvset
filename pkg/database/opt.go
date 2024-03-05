package database

import (
	"fmt"
	"time"
)

const (
	defaultDialect   = "postgres"
	defaultLocalHost = "@"
)

type Opt struct {
	Host               string        `yaml:"host"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	Port               string        `yaml:"port"`
	Name               string        `yaml:"name"`
	Dialect            string        `yaml:"dialect"`
	SQLitePath         string        `json:"sqlite_path"`
	Debug              bool          `yaml:"debug"`
	MaxIdleConns       int           `yaml:"max_idle_conns"`
	MaxOpenConns       int           `yaml:"max_open_conns"`
	MaxConnMaxLifetime time.Duration `yaml:"max_conn_max_lifetime"`
}

func (o *Opt) IsPostgres() bool {
	return o.Dialect == defaultDialect
}

func (o *Opt) IsSqlite3() bool {
	return o.Dialect == "sqlite3"
}

func (o *Opt) UnwrapOrPanic() {
	switch o.Dialect {
	case "":
		o.Dialect = defaultDialect
	case "postgres", "sqlite3":
		// do nothing
	default:
		panic("unsupported database dialect")
	}

	if o.Host == "" {
		o.Host = defaultLocalHost
	}

	if !o.IsSqlite3() {
		if o.MaxIdleConns <= 0 {
			panic("max_idle_conns must be greater than zero")
		}
		if o.MaxOpenConns <= 0 {
			panic("max_open_conns must be greater than zero")
		}
		if o.MaxConnMaxLifetime <= 0 {
			panic("max_conn_max_lifetime must be greater than zero")
		}
	}
}

// nolint:lll // it's okay
func (o *Opt) ConnectionString() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", o.User, o.Password, o.Host, o.Port, o.Name,
	)
}
