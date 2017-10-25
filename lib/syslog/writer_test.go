package syslog

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	ecslogs "github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/ecs-logs/lib"
)

const testGoroutines = 50

func TestWriter(t *testing.T) {
	// Start a bunch of workers who open and close writers like crazy.
	errc := make(chan error, testGoroutines)
	start := time.Now()
	for i := 0; i < testGoroutines; i++ {
		go func(i int) {
			for {
				d := time.Duration(rand.Intn(10)) * time.Millisecond
				time.Sleep(d)
				w, err := NewWriter("foo", "bar")
				if err != nil {
					errc <- err
				}
				err = w.WriteMessage(lib.Message{
					Group:  "foo",
					Stream: "bar",
					Event:  ecslogs.MakeEvent(ecslogs.INFO, fmt.Sprintf("slept %v", d)),
				})
				if err != nil {
					errc <- err
				}
				if err := w.Close(); err != nil {
					errc <- err
				}

				if time.Since(start) >= 10*time.Second {
					errc <- nil
					return
				}
			}
		}(i)
	}

	for i := 0; i < testGoroutines; i++ {
		if err := <-errc; err != nil {
			t.Error(err)
		}
	}

	//t.Logf("total new connections made: %d", atomic.LoadUint64(&newConnections))
}
