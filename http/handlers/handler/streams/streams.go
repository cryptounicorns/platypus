package streams

import (
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/corpix/formats"
	"github.com/corpix/loggers"
	"github.com/cryptounicorns/websocket"
	"github.com/cryptounicorns/websocket/writer"

	"github.com/cryptounicorns/platypus/http/handlers/consumer"
	"github.com/cryptounicorns/platypus/http/handlers/routers"
	"github.com/cryptounicorns/platypus/iopool"
)

type Streams struct {
	*websocket.UpgradeHandler
	config          Config
	log             loggers.Logger
	done            chan struct{}
	websocketFormat formats.Format
	Connections     *iopool.Writer
	Router          routers.Router
	Consumers       []*consumer.Stream
}

func (s *Streams) websocketWorker(wrap *template.Template, meta *consumer.Meta) {
	for {
		select {
		case <-s.done:
			return
		case r := <-meta.Stream:
			// FIXME: I don't like this error handling
			var (
				v   = r.Value
				buf []byte
				err error
			)

			if r.Err != nil {
				// XXX: Consumer always closes after error, so return here.
				s.log.Error(r.Err)
				return
			}

			buf, err = s.websocketFormat.Marshal(v)
			if err != nil {
				s.log.Error(r.Err)
				continue
			}

			err = wrap.Execute(
				s.Router,
				struct {
					Consumer consumer.Config
					Event    struct {
						JSON  []byte
						Value interface{}
					}
				}{
					Consumer: meta.Config,
					Event: struct {
						JSON  []byte
						Value interface{}
					}{
						JSON:  buf,
						Value: v,
					},
				},
			)
			if err != nil {
				s.log.Error(err)
				continue
			}
		}
	}
}

func (s *Streams) websocketWorkers(wrap *template.Template) {
	for _, cr := range s.Consumers {
		go s.websocketWorker(wrap, cr.Meta)
	}
}

func (s *Streams) ServeWebsocket(c io.WriteCloser, r *http.Request) {
	s.Connections.Add(writer.NewServerText(c))
}

func (s *Streams) Close() error {
	var (
		err error
	)

	close(s.done)

	// FIXME: Refactor writer pool to writecloser
	// So we could close all connections.
	for _, cr := range s.Consumers {
		err = cr.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func New(c Config, l loggers.Logger) (*Streams, error) {
	var (
		writerPool      = iopool.NewWriter()
		websocketFormat formats.Format
		consumers       []*consumer.Stream
		t               *template.Template
		r               routers.Router
		s               *Streams
		err             error
	)

	websocketFormat, err = formats.New(c.Format)
	if err != nil {
		return nil, err
	}

	t, err = template.New("wrap").Parse(
		strings.TrimSpace(c.Wrap),
	)
	if err != nil {
		return nil, err
	}

	r, err = routers.New(
		c.Router,
		writerPool,
		routers.NewWriterPoolCleaner(writerPool, l),
		l,
	)
	if err != nil {
		return nil, err
	}

	consumers, err = consumer.NewStreams(c.Consumers, l)
	if err != nil {
		return nil, err
	}

	s = &Streams{
		config:          c,
		log:             l,
		websocketFormat: websocketFormat,
		Connections:     writerPool,
		Router:          r,
		Consumers:       consumers,
	}

	s.UpgradeHandler = websocket.NewUpgradeHandler(s, l)

	s.websocketWorkers(t)

	return s, nil
}
