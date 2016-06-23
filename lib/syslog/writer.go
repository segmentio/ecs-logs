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
	DefaultTemplate = "<{{.PRIVAL}}>{{.TIMESTAMP}} {{.GROUP}}[{{.STREAM}}]: {{.MSG}}"
)

type DialConfig struct {
	Network    string
	Address    string
	Template   string
	TimeFormat string
}

func NewMessageBatchWriter(group string, stream string) (ecslogs.MessageBatchWriteCloser, error) {
	return DialWriter(DialConfig{})
}

func DialWriter(config DialConfig) (w ecslogs.MessageBatchWriteCloser, err error) {
	var netopts []string
	var addropts []string
	var conn net.Conn

	if len(config.Network) != 0 {
		netopts = []string{config.Network}
		addropts = []string{config.Address}
	} else {
		netopts = []string{"unixgram", "unix"}
		addropts = []string{"/dev/log", "/var/run/syslog", "/var/run/log"}
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

	w = NewWriter(WriterConfig{
		Backend:    conn,
		Template:   config.Template,
		TimeFormat: config.TimeFormat,
	})
	return
}

type WriterConfig struct {
	Backend    io.Writer
	Template   string
	TimeFormat string
}

func NewWriter(config WriterConfig) ecslogs.MessageBatchWriteCloser {
	if len(config.TimeFormat) == 0 {
		config.TimeFormat = time.Stamp
	}

	if len(config.Template) == 0 {
		config.Template = DefaultTemplate
	}

	return syslogWriter{
		WriterConfig: config,
		buf:          bufio.NewWriter(config.Backend),
		tpl:          newWriterTemplate(config.Template),
	}
}

func newWriterTemplate(format string) *template.Template {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	t := template.New("syslog")
	template.Must(t.Parse(format))
	return t
}

type syslogWriter struct {
	WriterConfig
	buf *bufio.Writer
	tpl *template.Template
}

func (w syslogWriter) Close() (err error) {
	if err = w.buf.Flush(); err != nil {
		return
	}

	if c, ok := w.Backend.(io.Closer); ok {
		err = c.Close()
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
		TIMESTAMP: msg.Time.Format(w.TimeFormat),
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

	if err = w.tpl.Execute(w.buf, m); err == nil {
		err = w.buf.Flush()
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
