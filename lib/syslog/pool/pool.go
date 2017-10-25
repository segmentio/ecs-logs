package pool

import (
	"io"
)

// A LimitedConnPool is a connection pool, with the property that
// the total number of live connections is limited to the size
// parameter passed to New.
type LimitedConnPool struct {
	conns  chan io.WriteCloser // Available connections in the pool
	live   chan struct{}       // Keep a count of living connections (in our hands, or the client)
	signal chan struct{}       // Used to wake up the connection producer
	err    chan error          // Send dial errors back to the client
}

// New returns a new LimitedConnPool.
func New(dial func() (io.WriteCloser, error), size int) (*LimitedConnPool, error) {
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
		err:    make(chan error),
	}

	p.conns <- w
	p.live <- struct{}{}

	// keep p.conns populated
	go func() {
		for range p.signal {
			for len(p.live) < size {
				w, err := dial()
				if err != nil {
					p.err <- err
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

// Put returns a connection to the pool.
func (p *LimitedConnPool) Put(w io.WriteCloser, dead bool) error {
	if dead {
		<-p.live               // decrement the live count
		p.signal <- struct{}{} // signal the connection dialer
		return w.Close()
	}
	p.conns <- w
	return nil
}

// Get gets a connection from the pool.
func (p *LimitedConnPool) Get() io.WriteCloser {
	return <-p.conns
}

// Errors returns a channel of errors encountered when dialing new
// connections. This channel must be consumed, or new connections
// will not be dialed. To limit the retry rate, limit the consumption rate
// of the error channel.
func (p *LimitedConnPool) Errors() <-chan error {
	return p.err
}
