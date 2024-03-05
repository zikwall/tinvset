package repositories

const sqlite3Schema = `
create table if not exists orderbooks (
    id integer primary key autoincrement,
    figi text,
    instrumentUid text,
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
const postgresSchema = `create table if not exists orderbooks (
    id serial primary key,
    figi text,
    instrument_uid text,
    depth integer,
    is_consistent integer,
    time_unix bigint,
    limit_up real,
    limit_down real
);

create table if not exists bids (
    id serial primary key,
    orderbook_id integer references orderbooks (id) on delete cascade,
    price real,
    quantity integer
);

create table if not exists asks (
    id serial primary key,
    orderbook_id integer references orderbooks (id) on delete cascade,
    price real,
    quantity integer
);`

func getSchema(by string) string {
	switch by {
	case "sqlite3":
		return sqlite3Schema
	}

	return postgresSchema
}
