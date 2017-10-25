package pool

import (
	"errors"
	"io"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	const size = 100

	// dialer which randomly fails 10% of the time
	dial := func() (io.WriteCloser, error) {
		if n := rand.Intn(10); n < 9 {
			return os.OpenFile(os.DevNull, os.O_RDWR, 0)
		}
		return nil, errors.New("failed to dial")
	}

	p, err := New(dial, size)
	if err != nil {
		t.Fatal(err)
	}

	// consume errors from the pool
	var failures uint64
	go func() {
		for range p.Errors() {
			atomic.AddUint64(&failures, 1)
		}
	}()

	// report pool stats once per second
	go func() {
		start := time.Now()
		for range time.Tick(1 * time.Second) {
			t.Logf("T+%02.2vs: live=%d pool=%d failures=%d\n", time.Since(start).Seconds(), len(p.live), len(p.conns), atomic.LoadUint64(&failures))
		}
	}()

	// start 2 * (pool size) writer goroutines
	for i := 0; i < 2*size; i++ {
		go func() {
			for {
				// wait a bit before getting a conn from the pool
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				w := p.Get()

				// random number of writes, with random waits in between
				for j := 0; j < rand.Intn(10); j++ {
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
					if _, err := w.Write([]byte("test")); err != nil {
						t.Error(err)
					}
				}

				// randomly decide if the connection should be considered dead
				dead := false
				if rand.Intn(10) == 9 {
					dead = true
				}

				// return the conn to the pool
				if err := p.Put(w, dead); err != nil {
					t.Error(err)
				}
			}
		}()
	}

	// let the test run for a bit
	if testing.Short() {
		time.Sleep(2 * time.Second)
	} else {
		time.Sleep(10 * time.Second)
	}
}
