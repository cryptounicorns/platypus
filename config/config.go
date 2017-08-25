package config

import (
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/corpix/formats"
	"github.com/corpix/pool"
	"github.com/corpix/queues"
	"github.com/corpix/queues/queue/nsq"
	"github.com/imdario/mergo"

	"github.com/cryptounicorns/market-fetcher-http/consumer"
	"github.com/cryptounicorns/market-fetcher-http/feeds"
	"github.com/cryptounicorns/market-fetcher-http/http"
	api "github.com/cryptounicorns/market-fetcher-http/http/api/config"
	"github.com/cryptounicorns/market-fetcher-http/logger"
	"github.com/cryptounicorns/market-fetcher-http/stores"
	"github.com/cryptounicorns/market-fetcher-http/stores/store/memory"
	"github.com/cryptounicorns/market-fetcher-http/transmitters"
	"github.com/cryptounicorns/market-fetcher-http/transmitters/transmitter/broadcast"
)

var (
	// LoggerConfig represents default logger config.
	LoggerConfig = logger.Config{
		Level: "info",
	}

	// FeedConfig represents default data sources config.
	FeedsConfig = feeds.Config{
		"tickers": queues.Config{
			Type: queues.NsqQueueType,
			Nsq: nsq.Config{
				Addr:     "127.0.0.1:4150",
				Topic:    "ticker",
				Channel:  "market-fetcher-http",
				LogLevel: nsq.LogLevelInfo,
			},
		},
	}

	// HTTPConfig represents default http server config.
	HTTPConfig = http.Config{
		Addr: ":8080",
		Api: api.Config{
			Consumer: consumer.Config{
				Format: "json",
			},
			Store: stores.Config{
				Type:   stores.MemoryStoreType,
				Memory: memory.Config{},
			},
			Transmitter: transmitters.Config{
				Type: transmitters.BroadcastTransmitterType,
				Broadcast: broadcast.Config{
					WriteTimeout: 10 * time.Second,
					Pool: pool.Config{
						Workers:   128,
						QueueSize: 1024,
					},
				},
			},
		},
	}

	// Default represents default application config.
	Default = Config{
		Logger: LoggerConfig,
		Feeds:  FeedsConfig,
		HTTP:   HTTPConfig,
	}
)

// Config represents application configuration structure.
type Config struct {
	Logger logger.Config
	Feeds  feeds.Config
	HTTP   http.Config
}

// FromReader fills Config structure `c` passed by reference with
// parsed config data in some `f` from reader `r`.
// It copies `Default` into the target structure before unmarshaling
// config, so it will have default values.
func FromReader(f formats.Format, r io.Reader, c *Config) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = mergo.Merge(c, Default)
	if err != nil {
		return err
	}

	return f.Unmarshal(data, c)
}

// FromFile fills Config structure `c` passed by reference with
// parsed config data from file in `path`.
func FromFile(path string, c *Config) error {
	f, err := formats.NewFromPath(path)
	if err != nil {
		return err
	}

	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()

	return FromReader(f, r, c)
}
