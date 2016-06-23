package syslog

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/segmentio/ecs-logs/lib"
)

const (
	DefaultFormat = "<{{.PRIVAL}}>{{.TIMESTAMP}} {{.GROUP}}[{{.STREAM}}]: {{.MSG}}"
)

func NewWriterTemplate(format string) *template.Template {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	t := template.New("syslog")
	template.Must(t.Parse(format))
	return t
}

func NewMessageBatchWriter(group string, stream string) (ecslogs.MessageBatchWriteCloser, error) {
	return DialWriter("", "", nil)
}

func DialWriter(network string, address string, template *template.Template) (w ecslogs.MessageBatchWriteCloser, err error) {
	var timeFormat string
	var netopts []string
	var addropts []string
	var conn net.Conn

	if len(network) != 0 {
		netopts = []string{network}
		addropts = []string{address}
		timeFormat = "2006-01-02T15:04:05.999Z07:00"
	} else {
		netopts = []string{"unixgram", "unix"}
		addropts = []string{"/dev/log", "/var/run/syslog", "/var/run/log"}
		timeFormat = time.Stamp
	}

connect:
	for _, n := range netopts {
		for _, a := range addropts {
			if conn, err = net.Dial(n, a); err == nil {
				break connect
			}
		}
	}

	if err != nil {
		return
	}

	w = NewWriter(conn, template, timeFormat)
	return
}

func NewWriter(w io.Writer, t *template.Template, timeFormat string) ecslogs.MessageBatchWriteCloser {
	if t == nil {
		t = NewWriterTemplate(DefaultFormat)
	}
	c, _ := w.(io.Closer)
	return syslogWriter{
		c: c,
		t: t,
		b: bufio.NewWriter(w),
		f: timeFormat,
	}
}

type syslogWriter struct {
	c io.Closer
	b *bufio.Writer
	t *template.Template
	f string
}

func (w syslogWriter) Close() (err error) {
	if err = w.b.Flush(); err != nil {
		return
	}

	if w.c != nil {
		err = w.c.Close()
	}

	return
}

func (w syslogWriter) WriteMessageBatch(batch []ecslogs.Message) (err error) {
	for _, msg := range batch {
		if err = w.WriteMessage(msg); err != nil {
			return
		}
	}
	return
}

func (w syslogWriter) WriteMessage(msg ecslogs.Message) (err error) {
	m := syslogMessage{
		PRIVAL:    int(msg.Level),
		HOSTNAME:  msg.Host,
		MSGID:     msg.ID,
		GROUP:     msg.Group,
		STREAM:    msg.Stream,
		MSG:       msg.Content,
		TIMESTAMP: msg.Time.Format(w.f),
	}

	if len(m.HOSTNAME) == 0 {
		m.HOSTNAME = "-"
	}

	if len(m.MSGID) == 0 {
		m.MSGID = "-"
	}

	if msg.PID == 0 {
		m.PROCID = "-"
	} else {
		m.PROCID = strconv.Itoa(msg.PID)
	}

	if len(msg.File) != 0 || len(msg.Func) != 0 {
		m.SOURCE = fmt.Sprintf("%s:%s:%d", msg.File, msg.Func, msg.Line)
	}

	if err = w.t.Execute(w.b, m); err == nil {
		err = w.b.Flush()
	}

	return
}

type syslogMessage struct {
	PRIVAL    int
	HOSTNAME  string
	PROCID    string
	MSGID     string
	GROUP     string
	STREAM    string
	MSG       string
	SOURCE    string
	TIMESTAMP string
}
