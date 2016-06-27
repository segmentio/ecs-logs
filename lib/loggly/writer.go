package loggly

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/segmentio/ecs-logs/lib"
	"github.com/segmentio/ecs-logs/lib/syslog"
)

func NewWriter(group string, stream string) (w ecslogs.Writer, err error) {
	var endpoint string
	var protocol string
	var address string
	var token string
	var pen string
	var tags string
	var template string
	var timeFormat string

	if endpoint, err = getEndpoint(); err != nil {
		return
	}

	if protocol, address, token, pen, tags, err = parseEndpoint(endpoint, group, stream); err != nil {
		return
	}

	if template = os.Getenv("LOGGLY_TEMPLATE"); len(template) == 0 {
		template = "<{{.PRIVAL}}>1 {{.TIMESTAMP}} {{.HOSTNAME}} {{.GROUP}} {{.PROCID}} {{.MSGID}} [{{.TAG}}] {{.MSG}}"
	}

	if timeFormat = os.Getenv("LOGGLY_TIME_FORMAT"); len(timeFormat) == 0 {
		timeFormat = "2006-01-02T15:04:05.999Z07:00"
	}

	return syslog.DialWriter(syslog.WriterConfig{
		Network:    protocol,
		Address:    address,
		Template:   template,
		TimeFormat: timeFormat,
		Tag:        fmt.Sprintf("%s@%s %s", token, pen, tags),
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
}

func getEndpoint() (endpoint string, err error) {
	var token string

	if endpoint = os.Getenv("LOGGLY_URL"); len(endpoint) != 0 {
		return
	}

	if token = os.Getenv("LOGGLY_TOKEN"); len(token) != 0 {
		endpoint = "tls://" + token + "@logs-01.loggly.com:6514"
		return
	}

	err = fmt.Errorf("missing LOGGLY_URL or LOGGLY_TOKEN environment variable")
	return
}

func parseEndpoint(endpoint string, group string, stream string) (protocol string, address string, token string, pen string, tags string, err error) {
	var u *url.URL
	var q url.Values

	if u, err = url.Parse(endpoint); err != nil {
		err = fmt.Errorf("invalid loggly endpoint, %s: %s", err, endpoint)
		return
	}

	if q, err = url.ParseQuery(u.RawQuery); err != nil {
		err = fmt.Errorf("invalid query string in loggly endpoint, %s: %s", err, endpoint)
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

	pen = extractPEN(u)
	tags = extractTags(u, q, group, stream)
	return
}

func extractProtocol(u *url.URL) (protocol string, err error) {
	switch u.Scheme {
	case "tcp", "tls":
		protocol = u.Scheme
	case "":
		err = fmt.Errorf("missing protocol in loggly endpoint: %s", u)
	default:
		err = fmt.Errorf("unsupported protocol in loggly endpoint, must be one of 'tcp' or 'tls': %s", u)
	}
	return
}

func extractAddress(u *url.URL) (address string, err error) {
	if len(u.Host) == 0 {
		err = fmt.Errorf("missing host in loggly endpoint: %s", u)
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
		err = fmt.Errorf("missing token parameter in loggly endpoint: %s", u)
	}

	return
}

func extractPEN(u *url.URL) (pen string) {
	if u.User != nil {
		pen, _ = u.User.Password()
	}

	if len(pen) == 0 {
		pen = "41058"
	}

	return
}

func extractTags(u *url.URL, q url.Values, group string, stream string) string {
	tags := q["tag"]
	list := make([]string, 0, len(tags)+2)
	list = append(list, group, stream)
	list = append(list, tags...)

	for i, tag := range list {
		list[i] = fmt.Sprintf("tag=%#v", tag)
	}

	return strings.Join(list, " ")
}
