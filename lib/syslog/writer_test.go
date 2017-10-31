package syslog

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/url"
	"os"
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
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				w, err := NewWriter("foo", "bar")
				if err != nil {
					errc <- err
				}
				for j := 0; j < rand.Intn(10); j++ {
					d := time.Duration(rand.Intn(30)) * time.Millisecond
					time.Sleep(d)
					err = w.WriteMessage(lib.Message{
						Group:  "foo",
						Stream: "bar",
						Event:  ecslogs.MakeEvent(ecslogs.INFO, fmt.Sprintf("slept %v", d)),
					})
					if err != nil {
						errc <- err
					}

				}

				if err := w.Close(); err != nil {
					errc <- err
				}

				if time.Since(start) >= 5*time.Second {
					// signal to the main goroutine that we
					// are exiting with no errors.
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
}

func TestGetPool(t *testing.T) {
	s := os.Getenv("SYSLOG_URL")
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal("SYSLOG_URL must be set")
	}

	network := u.Scheme
	address := u.Host

	opts := dialOpts{
		network: network,
		address: address,
		tls: &tls.Config{
			InsecureSkipVerify: true,
		},
		socksProxy: "",
	}
	pool1, err := getPool(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer pool1.Close()

	// make a new dialOpts identical to the old one
	opts = dialOpts{
		network: network,
		address: address,
		tls: &tls.Config{
			InsecureSkipVerify: true,
		},
		socksProxy: "",
	}
	pool2, err := getPool(opts)
	if err != nil {
		t.Fatal(err)
	}

	if pool1 != pool2 {
		t.Error("pools did not match")
	}
}
