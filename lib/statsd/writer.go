package statsd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/ecs-logs/lib"
	"github.com/statsd/client"
)

type Client interface {
	io.Closer

	Flush() error

	IncrEvents(ecslogs.Level, int) error
}

type WriterConfig struct {
	Address string
	Group   string
	Stream  string
	Dial    func(addr string, group string, stream string) (Client, error)
}

func NewWriter(group string, stream string) (w lib.Writer, err error) {
	var c WriterConfig
	var s string
	var u *url.URL

	if s = os.Getenv("STATSD_URL"); len(s) != 0 {
		if u, err = url.Parse(s); err != nil {
			err = fmt.Errorf("invalid statsd URL: %s", err)
			return
		}

		if u.Scheme != "udp" {
			err = fmt.Errorf("invalid statsd URL: only the UDP protocol is supported but %s was found", u.Scheme)
			return
		}

		c.Address = u.Host
	}

	c.Group = group
	c.Stream = stream

	return DialWriter(c)
}

func DialWriter(config WriterConfig) (w lib.Writer, err error) {
	var client Client

	if len(config.Address) == 0 {
		config.Address = "localhost:8125"
	}

	if config.Dial == nil {
		config.Dial = dial
	}

	if client, err = config.Dial(config.Address, config.Group, config.Stream); err != nil {
		return
	}

	w = writer{client}
	return
}

func dial(addr string, group string, stream string) (Client, error) {
	if cli, err := statsd.Dial(addr); err != nil {
		return nil, err
	} else {
		cli.Prefix("ecs-logs." + group + ".")
		return client{cli}, nil
	}
}

type client struct {
	*statsd.Client
}

func (c client) IncrEvents(level ecslogs.Level, value int) error {
	return c.IncrBy(strings.ToLower(level.String()), value)
}

type writer struct {
	client Client
}

type metric struct {
	value int
}

func (w writer) Close() error {
	return w.client.Close()
}

func (w writer) WriteMessage(msg lib.Message) error {
	return w.WriteMessageBatch(lib.MessageBatch{msg})
}

func (w writer) WriteMessageBatch(batch lib.MessageBatch) error {
	return sendMetrics(w.client, extractMetrics(batch))
}

func extractMetrics(batch lib.MessageBatch) map[ecslogs.Level]*metric {
	metrics := make(map[ecslogs.Level]*metric, 10)

	for _, msg := range batch {
		if m := metrics[msg.Event.Level]; m == nil {
			m = &metric{value: 1}
			metrics[msg.Event.Level] = m
		} else {
			m.value++
		}
	}

	return metrics
}

func sendMetrics(client Client, metrics map[ecslogs.Level]*metric) (err error) {
	for lvl, met := range metrics {
		if e := client.IncrEvents(lvl, met.value); e != nil {
			err = lib.AppendError(err, e)
		}
	}
	if e := client.Flush(); e != nil {
		err = lib.AppendError(err, e)
	}
	return
}
