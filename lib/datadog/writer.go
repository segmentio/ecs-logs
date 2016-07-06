package datadog

import (
	"fmt"
	"net/url"
	"os"

	"github.com/segmentio/ecs-logs/lib"
	"github.com/segmentio/ecs-logs/lib/statsd"
	"github.com/statsd/datadog"
)

func NewWriter(group string, stream string) (w lib.Writer, err error) {
	var c statsd.WriterConfig
	var s string
	var u *url.URL

	if s = os.Getenv("DATADOG_URL"); len(s) != 0 {
		if u, err = url.Parse(s); err != nil {
			err = fmt.Errorf("invalid datadog URL: %s", err)
			return
		}

		if u.Scheme != "udp" {
			err = fmt.Errorf("invalid datadog URL: only the UDP protocol is supported but %s was found", u.Scheme)
			return
		}

		c.Address = u.Host
	}

	c.Group = group
	c.Stream = stream
	c.Dial = dialUdpClient

	return statsd.DialWriter(c)
}

type client struct {
	*datadog.Client
}

func dialUdpClient(addr string, group string, stream string) (statsd.Client, error) {
	if dd, err := datadog.Dial(addr); err != nil {
		return nil, err
	} else {
		dd.SetPrefix("ecs-logs." + group + ".")
		dd.SetTags("group:"+group, "stream:"+stream)
		return client{dd}, nil
	}
}

func (c client) IncrBy(name string, value int) error {
	return c.Client.IncrBy(name, value)
}
