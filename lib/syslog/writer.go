package syslog

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/segmentio/ecs-logs/lib"
	"github.com/segmentio/ecs-logs/lib/syslog/pool"

	"golang.org/x/net/proxy"
)

const DefaultTemplate = "<{{.PRIVAL}}>{{.TIMESTAMP}} {{.GROUP}}[{{.STREAM}}]: {{.MSG}}"

const poolSize = 20

var (
	connPoolsLock sync.Mutex
	connPools     map[dialOpts]*pool.LimitedConnPool
)

type WriterConfig struct {
	Network    string
	Address    string
	Template   string
	TimeFormat string
	Tag        string
	TLS        *tls.Config
	SocksProxy string
}

type dialOpts struct {
	network    string
	address    string
	tls        *tls.Config
	socksProxy string
}

func init() {
	connPools = make(map[dialOpts]*pool.LimitedConnPool)
}

func NewWriter(group, stream string) (lib.Writer, error) {
	var c WriterConfig

	if s := os.Getenv("SYSLOG_URL"); len(s) != 0 {
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid syslog URL: %s", err)
		}

		c.Network = u.Scheme
		c.Address = u.Host
	}

	c.Template = os.Getenv("SYSLOG_TEMPLATE")
	c.TimeFormat = os.Getenv("SYSLOG_TIME_FORMAT")

	return DialWriter(c)
}

func DialWriter(config WriterConfig) (lib.Writer, error) {
	var netopts, addropts []string

	if len(config.Network) != 0 {
		netopts = []string{config.Network}
	} else if len(config.Address) == 0 || strings.HasPrefix(config.Address, "/") {
		// When starting with a '/' we assume it's gonna be a file path,
		// otherwise we fallback to trying a TLS connection so we don't
		// implicitly send logs over an unsecured link.
		config.Network = "unix"
		netopts = []string{"unixgram", "unix"}
	} else {
		netopts = []string{"tls"}
	}

	if len(config.Address) != 0 {
		addropts = []string{config.Address}
	} else if strings.HasPrefix(config.Network, "unix") {
		// This was copied from the standard log/syslog package, they do the same
		// and try to guess at runtime which socket syslogd is using.
		addropts = []string{"/dev/log", "/var/run/syslog", "/var/run/log"}
	} else {
		// The config doesn't point to a unix domain socket, falling back to trying
		// to connect to syslogd over a network interface.
		addropts = []string{"localhost:514"}
	}

	// Try various fallbacks if no hints were given
	var w *writer
	var err error
	for _, n := range netopts {
		for _, a := range addropts {
			opts := dialOpts{
				network:    n,
				address:    a,
				tls:        config.TLS,
				socksProxy: config.SocksProxy,
			}
			if w, err = newWriter(opts, config); err == nil {
				return w, nil
			}
		}
	}

	return nil, err
}

type writer struct {
	// configuration
	timefmt string
	tpl     *template.Template
	tag     string

	// connection state
	pool    *pool.LimitedConnPool
	backend io.WriteCloser
	dead    bool

	// buffered i/o
	buf   bytes.Buffer
	out   func(*writer, message) error
	flush func() error
}

func newWriter(opts dialOpts, cfg WriterConfig) (*writer, error) {
	var out func(*writer, message) error
	var flush func() error

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.Stamp
	}

	if cfg.Template == "" {
		cfg.Template = DefaultTemplate
	}

	p, err := getPool(opts)
	if err != nil {
		return nil, err
	}

	backend := p.Get()
	switch b := backend.(type) {
	case bufferedWriter:
		out, flush = (*writer).directWrite, b.Flush
	default:
		out, flush = (*writer).bufferedWrite, func() error { return nil }
	}

	return &writer{
		timefmt: cfg.TimeFormat,
		tpl:     newWriterTemplate(cfg.Template),
		tag:     cfg.Tag,

		backend: backend,
		pool:    p,

		flush: flush,
		out:   out,
	}, nil
}

// getPool returns a connection pool for the given configuration.
func getPool(opts dialOpts) (*pool.LimitedConnPool, error) {
	connPoolsLock.Lock()
	defer connPoolsLock.Unlock()

	p, ok := connPools[opts]
	if !ok {
		// dial closes over opts
		dial := func() (io.WriteCloser, error) {
			//fmt.Fprintln(os.Stderr, "dialing new connection")
			return dialWriter(opts.network, opts.address, opts.tls, opts.socksProxy)
		}
		var err error
		p, err = pool.New(dial, poolSize)
		if err != nil {
			return nil, err
		}
		connPools[opts] = p
	}

	return p, nil
}

func newWriterTemplate(format string) *template.Template {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	t := template.New("syslog")
	template.Must(t.Parse(format))
	return t
}

func (w *writer) Close() (err error) {
	return w.pool.Put(w.backend, w.dead)
}

func (w *writer) WriteMessageBatch(batch lib.MessageBatch) error {
	for _, msg := range batch {
		if err := w.write(msg); err != nil {
			w.dead = true
			return err
		}
	}
	if err := w.flush(); err != nil {
		w.dead = true
		return err
	}
	return nil
}

func (w *writer) WriteMessage(msg lib.Message) error {
	if err := w.write(msg); err != nil {
		w.dead = true
		return err
	}
	if err := w.flush(); err != nil {
		w.dead = true
		return err
	}

	return nil
}

func (w *writer) write(msg lib.Message) (err error) {
	m := message{
		PRIVAL:    int(msg.Event.Level-1) + 8, // +8 is for user-level messages facility
		HOSTNAME:  msg.Event.Info.Host,
		MSGID:     msg.Event.Info.ID,
		GROUP:     msg.Group,
		STREAM:    msg.Stream,
		TIMESTAMP: msg.Event.Time.Format(w.timefmt),
		TAG:       w.tag,
	}

	if len(m.HOSTNAME) == 0 {
		m.HOSTNAME = "-"
	}

	if len(m.MSGID) == 0 {
		m.MSGID = "-"
	}

	if msg.Event.Info.PID == 0 {
		m.PROCID = "-"
	} else {
		m.PROCID = strconv.Itoa(msg.Event.Info.PID)
	}

	m.MSG = msg.Event.String()
	return w.out(w, m)
}

func (w *writer) directWrite(m message) (err error) {
	return w.tpl.Execute(w.backend, m)
}

func (w *writer) bufferedWrite(m message) (err error) {
	w.buf.Reset()
	w.tpl.Execute(&w.buf, m)
	_, err = w.backend.Write(w.buf.Bytes())
	return
}

type message struct {
	PRIVAL    int
	HOSTNAME  string
	PROCID    string
	MSGID     string
	GROUP     string
	STREAM    string
	TAG       string
	MSG       string
	TIMESTAMP string
}

type bufferedWriter interface {
	Flush() error
}

type bufferedConn struct {
	buf  *bufio.Writer
	conn net.Conn
}

func (c bufferedConn) Close() error                { return c.conn.Close() }
func (c bufferedConn) Flush() error                { return c.buf.Flush() }
func (c bufferedConn) Write(b []byte) (int, error) { return c.buf.Write(b) }

func dialWriter(network, address string, config *tls.Config, socksProxy string) (w io.WriteCloser, err error) {
	var conn, rawConn net.Conn
	var dial func(string, string) (net.Conn, error)
	var socksDialer proxy.Dialer

	if network == "tls" {
		network, dial = "tcp", func(network, address string) (net.Conn, error) {
			return tls.Dial(network, address, config)
		}
	} else {
		dial = net.Dial
	}

	if socksProxy != "" {
		if socksDialer, err = proxy.SOCKS5(network, socksProxy, nil, proxy.Direct); err != nil {
			return
		}

		dial = func(network, address string) (conn net.Conn, err error) {
			if config == nil {
				conn, err = socksDialer.Dial(network, address)
			} else {
				rawConn, err = socksDialer.Dial(network, address)
				if err != nil {
					return nil, err
				}

				tlsConn := tls.Client(rawConn, config)
				if err = tlsConn.Handshake(); err == nil {
					conn = tlsConn
				}
			}
			return
		}
	}

	for attempt := 1; true; attempt++ {
		if conn, err = dial(network, address); err == nil {
			break
		}

		if attempt == 3 {
			return
		}

		err = nil
		time.Sleep(1 * time.Second)
	}

	if err == nil {
		switch network {
		case "udp", "udp4", "udp6", "unixgram", "unixpacket":
			w = conn
		default:
			w = bufferedConn{
				conn: conn,
				buf:  bufio.NewWriter(conn),
			}
		}
	}

	return
}
