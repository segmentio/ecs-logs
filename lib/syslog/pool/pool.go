package pool

import (
	"io"
)

// A LimitedConnPool is a connection pool, with the property that
// the total number of live connections is limited to the size
// parameter passed to NewLimited. In order to make this guarantee,
// we assume that all new connections are introduced by calls to the
// Get method, and return to the pool by the Put method (whether
// or not the connection is dead)
type LimitedConnPool struct {
	conns  chan io.WriteCloser // Available connections in the pool
	live   chan struct{}       // Keep a count of living connections (in our hands, or the client)
	signal chan struct{}       // Used to wake up the connection producer
	err    chan error          // Send dial errors back to the client
}

// NewLimited returns a new LimitedConnPool with the given size limit and dial function.
func NewLimited(size int, dial func() (io.WriteCloser, error)) (*LimitedConnPool, error) {
	// Tentative first try - if this doesn't work, we assume it never will
	// and fail to initialize.
	w, err := dial()
	if err != nil {
		return nil, err
	}

	p := LimitedConnPool{
		conns:  make(chan io.WriteCloser, size),
		live:   make(chan struct{}, size),
		signal: make(chan struct{}),

		// try to make this large enough to avoid dropping
		// errors if clients only check errors occasionally
		err: make(chan error, size),
	}

	p.conns <- w
	p.live <- struct{}{}

	// keep p.conns populated
	go func() {
		for range p.signal {
			for len(p.live) < size {
				w, err := dial()
				if err != nil {
					select {
					case p.err <- err:
					default:
						// error channel is full, drop this error
					}
					continue
				}
				p.conns <- w
				p.live <- struct{}{}
			}
		}
	}()

	// kick off the producer
	p.signal <- struct{}{}

	return &p, nil
}

// Put returns a connection to the pool. If dead is true, the
// connection is removed from the pool so that a new connection
// can be dialed.
func (p *LimitedConnPool) Put(w io.WriteCloser, dead bool) error {
	if dead {
		<-p.live               // decrement the live count
		p.signal <- struct{}{} // signal the connection dialer
		return w.Close()
	}
	p.conns <- w
	return nil
}

// Get retrieves a connection from the pool, if available.
// A new connection will only be dialed if the total number
// of live connections is below the configured size limit.
func (p *LimitedConnPool) Get() io.WriteCloser {
	return <-p.conns
}

// Errors returns a channel of errors encountered when dialing new
// connections. Errors will be dropped if this channel is not
// consumed.
func (p *LimitedConnPool) Errors() <-chan error {
	return p.err
}
