package pool

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	const poolSize = 100

	dir, err := ioutil.TempDir("", "pool_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var newConnections int
	dial := func() (io.WriteCloser, error) {
		newConnections++
		return ioutil.TempFile(dir, "")
	}

	p, err := NewLimited(poolSize, dial)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// consume errors from the pool
	var failures uint64
	go func() {
		for err := range p.Errors() {
			atomic.AddUint64(&failures, 1)
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}()

	// report pool stats once per second
	go func() {
		start := time.Now()
		for range time.Tick(1 * time.Second) {
			t.Logf("T+%02.2vs: live=%d pool=%d failures=%d\n", time.Since(start).Seconds(), len(p.live), len(p.conns), atomic.LoadUint64(&failures))
		}
	}()

	var wg sync.WaitGroup

	// start 2*poolSize writer goroutines
	for i := 0; i < 2*poolSize; i++ {
		start := time.Now()
		stop := 10 * time.Second
		if testing.Short() {
			stop = 2 * time.Second
		}
		wg.Add(1)
		go func() {
			for time.Since(start) < stop {
				// wait a bit before getting a conn from the pool
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				w := p.Get()

				// random number of writes, with random waits in between
				for j := 0; j < rand.Intn(10); j++ {
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
					if _, err := w.Write([]byte("test\n")); err != nil {
						t.Error(err)
					}
				}

				// return the conn to the pool
				if err := w.Close(); err != nil {
					t.Error(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if newConnections != poolSize {
		t.Errorf("dialed %d connections, want %d", newConnections, poolSize)
	}
}
