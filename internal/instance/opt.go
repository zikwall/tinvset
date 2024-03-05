package instance

import (
	"github.com/tinkoff/invest-api-go-sdk/investgo"

	"github.com/bataloff/tiknkoff/pkg/database"
)

type Options struct {
	Database *database.Opt
	Invest   investgo.Config
}
