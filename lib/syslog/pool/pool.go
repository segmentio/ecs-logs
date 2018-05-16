package pool

import (
	"io"
	"time"

	"github.com/jpillora/backoff"
)

// A LimitedConnPool is a connection pool, with the property that
// the total number of live connections is limited to the size
// parameter passed to NewLimited. In order to make this guarantee,
// we assume that all new connections are introduced by calls to the
// Get method, and returned to the pool by calls to Close().
type LimitedConnPool struct {
	conns  chan *conn    // Available connections in the pool
	live   chan struct{} // Keep a count of living connections (in our hands, or the client)
	signal chan struct{} // Used to wake up the connection producer
	err    chan error    // Send dial errors back to the client
}

// conn wraps an io.WriteCloser, marking the connection as dead
// on any write error and changing the meaning of Close.
// Closing a conn calls the io.WriteCloser's Close method if
// the conn is marked dead, or returns it to the pool otherwise.
type conn struct {
	conn io.WriteCloser
	pool *LimitedConnPool
	dead bool
}

func (w *conn) Write(p []byte) (int, error) {
	n, err := w.conn.Write(p)
	if err != nil {
		w.dead = true
	}
	return n, err
}

func (w *conn) Close() error {
	return w.pool.put(w)
}

type bufferedWriter interface {
	Flush() error
}

func (w *conn) Flush() error {
	if t, ok := w.conn.(bufferedWriter); ok {
		return t.Flush()
	}
	return nil
}

// NewLimited returns a new LimitedConnPool with the given size limit and dial function.
func NewLimited(size int, dial func() (io.WriteCloser, error)) (*LimitedConnPool, error) {
	// Tentative first try - if this doesn't work, we assume it never will
	// and fail to initialize. This is admittedly not great, but we rely on
	// unreachable addresses failing immediately in our syslog package, which,
	// if no address is specified, attempts a number of fallback addresses
	// until one succeeds.
	w, err := dial()
	if err != nil {
		return nil, err
	}

	p := LimitedConnPool{
		conns:  make(chan *conn, size),
		live:   make(chan struct{}, size),
		signal: make(chan struct{}),

		// try to make this large enough to avoid dropping
		// errors if clients only check errors occasionally
		err: make(chan error, size),
	}

	p.conns <- &conn{
		conn: w,
		pool: &p,
	}
	p.live <- struct{}{}

	// keep p.conns populated
	go func() {
		// TODO: it would be nice if the client could control
		// backoff, but doing it here seems sufficient for now.
		backoff := &backoff.Backoff{
			Factor: 2,
			Min:    10 * time.Millisecond,
			Max:    10 * time.Second,
		}
		for range p.signal {
			for len(p.live) < size {
				w, err := dial()
				if err != nil {
					select {
					case p.err <- err:
					default:
						// error channel is full, drop this error.
					}
					time.Sleep(backoff.Duration())
					continue
				}
				backoff.Reset()
				p.conns <- &conn{
					conn: w,
					pool: &p,
				}
				p.live <- struct{}{}
			}
		}
	}()

	// kick off the producer
	p.signal <- struct{}{}

	return &p, nil
}

func (p *LimitedConnPool) Close() {
	// Important to close this first, so the dialer doesn't loop again.
	close(p.signal)

	// Close all the underlying connections
	close(p.conns)
	for c := range p.conns {
		c.conn.Close()
	}

	// Close the error channel, allowing any client error handling
	// range loops to finish
	close(p.err)
}

// put returns a connection to the pool. If the connection is dead,
// it is removed from the pool so that a new connection can be dialed.
func (p *LimitedConnPool) put(w *conn) error {
	if w.dead {
		// decrement the live connection count
		<-p.live

		// signal the connection dialer if necessary
		select {
		case p.signal <- struct{}{}:
		default:
		}

		return w.conn.Close()
	}
	p.conns <- w
	return nil
}

// Get retrieves a connection from the pool, if available.
// A new connection will only be dialed if the total number
// of live connections is below the configured size limit.
// Closing the returned io.WriteCloser automatically returns
// the connection to the pool.
func (p *LimitedConnPool) Get() io.WriteCloser {
	return <-p.conns
}

// Errors returns a channel of errors encountered when dialing new
// connections. Errors will be dropped if this channel is not
// consumed.
func (p *LimitedConnPool) Errors() <-chan error {
	return p.err
}
