package stream

import (
	"github.com/cryptounicorns/platypus/http/handlers/consumer"
	"github.com/cryptounicorns/platypus/http/handlers/routers"
)

type Config struct {
	Format string

	Consumer consumer.Config
	Router   routers.Config
}
