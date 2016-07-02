package statsd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/segmentio/ecs-logs/lib"
	"github.com/statsd/client"
)

type Client interface {
	io.Closer

	Flush() error

	IncrBy(name string, value int) error
}

type WriterConfig struct {
	Address string
	Group   string
	Stream  string
	Dial    func(addr string, group string, stream string) (Client, error)
}

func NewWriter(group string, stream string) (w ecslogs.Writer, err error) {
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

func DialWriter(config WriterConfig) (w ecslogs.Writer, err error) {
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
		return cli, nil
	}
}

type writer struct {
	client Client
}

type metric struct {
	name  string
	value int
}

func (w writer) Close() error {
	return w.client.Close()
}

func (w writer) WriteMessage(msg ecslogs.Message) error {
	return w.WriteMessageBatch([]ecslogs.Message{msg})
}

func (w writer) WriteMessageBatch(batch []ecslogs.Message) error {
	return sendMetrics(w.client, extractMetrics(batch))
}

func extractMetrics(batch []ecslogs.Message) map[ecslogs.Level]*metric {
	metrics := make(map[ecslogs.Level]*metric, 10)

	for _, msg := range batch {
		m := metrics[msg.Event.Level]

		if m == nil {
			m = &metric{name: strings.ToLower(msg.Event.Level.String())}
			metrics[msg.Event.Level] = m
		}

		m.value++
	}

	return metrics
}

func sendMetrics(client Client, metrics map[ecslogs.Level]*metric) (err error) {
	for _, m := range metrics {
		if e := client.IncrBy(m.name, m.value); e != nil {
			err = ecslogs.AppendError(err, e)
		}
	}
	if e := client.Flush(); e != nil {
		err = ecslogs.AppendError(err, e)
	}
	return
}
