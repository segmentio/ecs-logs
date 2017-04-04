package logdna

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/segmentio/ecs-logs/lib"
	"github.com/segmentio/ecs-logs/lib/syslog"
)

func NewWriter(group string, stream string) (w lib.Writer, err error) {
	var endpoint string
	var protocol string
	var address string
	var token string
	var tags string
	var template string
	var timeFormat string
	var socksProxy string

	if endpoint, err = getEndpoint(); err != nil {
		return
	}

	if protocol, address, token, tags, err = parseEndpoint(endpoint, group, stream); err != nil {
		return
	}

	if template = os.Getenv("LOGDNA_TEMPLATE"); len(template) == 0 {
		template = "<{{.PRIVAL}}>1 {{.TIMESTAMP}} {{.HOSTNAME}} {{.GROUP}} {{.STREAM}} {{.MSGID}} [{{.TAG}}] {{.MSG}}"
		if len(token) != 0 {
			template = "<key:" + token + "> " + template
		}
	}

	if timeFormat = os.Getenv("LOGDNA_TIME_FORMAT"); len(timeFormat) == 0 {
		timeFormat = "2016-02-10T09:28:01.982-08:00"
	}

	socksProxy = os.Getenv("SOCKS_PROXY")
	if _, _, err = net.SplitHostPort(socksProxy); err != nil {
		socksProxy = ""
	}

	return syslog.DialWriter(syslog.WriterConfig{
		Network:    protocol,
		Address:    address,
		Template:   template,
		TimeFormat: timeFormat,
		Tag:        fmt.Sprintf("logdna@48950 %s", tags),
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
		SocksProxy: socksProxy,
	})
}

func getEndpoint() (endpoint string, err error) {
	var token string

	if endpoint = os.Getenv("LOGDNA_URL"); len(endpoint) != 0 {
		return
	}

	if token = os.Getenv("LOGDNA_TOKEN"); len(token) != 0 {
		endpoint = "tls://syslog-a.logdna.com:6514"
		return
	}

	err = fmt.Errorf("missing LOGDNA_URL or LOGDNA_TOKEN environment variable")
	return
}

func parseEndpoint(endpoint string, group string, stream string) (protocol string, address string, token string, tags string, err error) {
	var u *url.URL
	var q url.Values

	if u, err = url.Parse(endpoint); err != nil {
		err = fmt.Errorf("invalid logdna endpoint, %s: %s", err, endpoint)
		return
	}

	if q, err = url.ParseQuery(u.RawQuery); err != nil {
		err = fmt.Errorf("invalid query string in logdna endpoint, %s: %s", err, endpoint)
		return
	}

	if protocol, err = extractProtocol(u); err != nil {
		return
	}

	if address, err = extractAddress(u); err != nil {
		return
	}

	if token, err = extractToken(u); err != nil {
		return
	}
	tags = extractTags(u, q)
	return
}

func extractProtocol(u *url.URL) (protocol string, err error) {
	switch u.Scheme {
	case "tcp", "tls":
		protocol = u.Scheme
	case "":
		err = fmt.Errorf("missing protocol in logdna endpoint: %s", u)
	default:
		err = fmt.Errorf("unsupported protocol in logdna endpoint, must be one of 'tcp' or 'tls': %s", u)
	}
	return
}

func extractAddress(u *url.URL) (address string, err error) {
	if len(u.Host) == 0 {
		err = fmt.Errorf("missing host in logdna endpoint: %s", u)
	} else {
		address = u.Host
	}
	return
}

func extractToken(u *url.URL) (token string, err error) {
	if u.User != nil {
		token = u.User.Username()
	}

	if len(token) == 0 {
		err = fmt.Errorf("missing token parameter in logdna endpoint: %s", u)
	}

	return
}

func extractTags(u *url.URL, q url.Values) string {
	return fmt.Sprintf("tags=\"%s\"", strings.Join(q["tag"], ","))
}
