package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/ecs-logs/lib"

	_ "github.com/segmentio/ecs-logs/lib/cloudwatchlogs"
	_ "github.com/segmentio/ecs-logs/lib/syslog"
)

func main() {
	var err error
	var src string
	var dst string
	var maxBytes int
	var maxCount int
	var flushTimeout time.Duration
	var cacheTimeout time.Duration

	flag.StringVar(&src, "src", "stdin", "The log source from which messages will be read ["+strings.Join(ecslogs.SourcesAvailable(), ", ")+"]")
	flag.StringVar(&dst, "dst", "stdout", "The log destination to which messages will be written ["+strings.Join(ecslogs.DestinationsAvailable(), ", ")+"]")
	flag.IntVar(&maxBytes, "max-batch-bytes", 1000000, "The maximum size in bytes of a message batch")
	flag.IntVar(&maxCount, "max-batch-size", 10000, "The maximum number of messages in a batch")
	flag.DurationVar(&flushTimeout, "flush-timeout", 5*time.Second, "How often messages will be flushed")
	flag.DurationVar(&cacheTimeout, "cache-timeout", 5*time.Minute, "How to wait before clearing unused internal cache")
	flag.Parse()

	var store = ecslogs.NewStore()
	var source ecslogs.Source
	var dest ecslogs.Destination
	var reader ecslogs.MessageReadCloser

	if source = ecslogs.GetSource(src); source == nil {
		fatalf("unknown log source (%s)", src)
	}

	if dest = ecslogs.GetDestination(dst); dest == nil {
		fatalf("unknown log destination (%s)", dst)
	}

	if reader, err = source.Open(); err != nil {
		fatalf("failed to open log source (%s)", err)
	}

	join := &sync.WaitGroup{}

	limits := ecslogs.StreamLimits{
		MaxCount: maxCount,
		MaxBytes: maxBytes,
		MaxTime:  flushTimeout,
	}

	expchan := time.Tick(flushTimeout)
	msgchan := make(chan ecslogs.Message, 1)
	sigchan := make(chan os.Signal, 1)

	go read(reader, msgchan)

	for {
		select {
		case msg, ok := <-msgchan:
			now := time.Now()

			if !ok {
				logf("reached EOF, waiting for all write operations to complete...")
				limits.Force = true
				flushAll(dest, store, limits, now, join)
				join.Wait()
				logf("done")
				return
			}

			group, stream := store.Add(msg, now)
			flush(dest, group, stream, limits, now, join)

		case <-expchan:
			logf("timer pulse, flushing streams that haven't been in a while and clearing expired cache...")
			now := time.Now()
			flushAll(dest, store, limits, now, join)
			store.RemoveExpired(cacheTimeout, now)

		case <-sigchan:
			logf("got interrupt signal, closing message reader...")
			reader.Close()
		}
	}
}

func read(r ecslogs.MessageReader, c chan<- ecslogs.Message) {
	defer close(c)

	hostname, _ := os.Hostname()

	for {
		if msg, err := r.ReadMessage(); err != nil {
			if err == io.EOF {
				break
			}
			errorf("the message reader failed (%s)", err)
		} else {
			if len(msg.Host) == 0 {
				msg.Host = hostname
			}

			if msg.Time == (time.Time{}) {
				msg.Time = time.Now()
			}

			c <- msg
		}
	}
}

func write(dest ecslogs.Destination, group string, stream string, batch []ecslogs.Message, join *sync.WaitGroup) {
	defer join.Done()

	var w ecslogs.MessageBatchWriteCloser
	var err error

	if w, err = dest.Open(group, stream); err != nil {
		errorf("dropping message batch of %d messages to %s::%s (%s)", len(batch), group, stream, err)
		return
	}
	defer w.Close()

	if err = w.WriteMessageBatch(batch); err != nil {
		errorf("dropping message batch of %d messages to %s::%s (%s)", len(batch), group, stream, err)
		return
	}
}

func flush(dest ecslogs.Destination, group *ecslogs.Group, stream *ecslogs.Stream, limits ecslogs.StreamLimits, now time.Time, join *sync.WaitGroup) {
	if batch := stream.Flush(limits, now); len(batch) != 0 {
		logf("flushing %d messages to %s::%s", len(batch), group.Name(), stream.Name())
		join.Add(1)
		go write(dest, group.Name(), stream.Name(), batch, join)
	}
}

func flushAll(dest ecslogs.Destination, store *ecslogs.Store, limits ecslogs.StreamLimits, now time.Time, join *sync.WaitGroup) {
	store.ForEach(func(group *ecslogs.Group) {
		group.ForEach(func(stream *ecslogs.Stream) {
			flush(dest, group, stream, limits, now, join)
		})
	})
}

func fatalf(format string, args ...interface{}) {
	errorf(format, args...)
	os.Exit(1)
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
