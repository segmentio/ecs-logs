package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/ecs-logs/lib"

	_ "github.com/segmentio/ecs-logs/lib/cloudwatchlogs"
	_ "github.com/segmentio/ecs-logs/lib/loggly"
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

	flag.StringVar(&src, "src", "stdin", "A comma separated list of log sources from which messages will be read ["+strings.Join(ecslogs.SourcesAvailable(), ", ")+"]")
	flag.StringVar(&dst, "dst", "stdout", "A comma separated list of log destinations to which messages will be written ["+strings.Join(ecslogs.DestinationsAvailable(), ", ")+"]")
	flag.IntVar(&maxBytes, "max-batch-bytes", 1000000, "The maximum size in bytes of a message batch")
	flag.IntVar(&maxCount, "max-batch-size", 10000, "The maximum number of messages in a batch")
	flag.DurationVar(&flushTimeout, "flush-timeout", 5*time.Second, "How often messages will be flushed")
	flag.DurationVar(&cacheTimeout, "cache-timeout", 5*time.Minute, "How to wait before clearing unused internal cache")
	flag.Parse()

	var store = ecslogs.NewStore()
	var sources []ecslogs.Source
	var dests []ecslogs.Destination
	var readers []ecslogs.Reader

	if sources = ecslogs.GetSources(strings.Split(src, ",")...); len(sources) == 0 {
		fatalf("no or invalid log sources")
	}

	if dests = ecslogs.GetDestinations(strings.Split(dst, ",")...); len(dests) == 0 {
		fatalf("no or invalid log destinations")
	}

	if readers, err = openSources(sources); err != nil {
		fatalf("failed to open sources (%s)", err)
	}

	join := &sync.WaitGroup{}

	limits := ecslogs.StreamLimits{
		MaxCount: maxCount,
		MaxBytes: maxBytes,
		MaxTime:  flushTimeout,
	}

	expchan := time.Tick(flushTimeout)
	msgchan := make(chan ecslogs.Message, len(readers))
	sigchan := make(chan os.Signal, 1)
	counter := int32(len(readers))
	startReaders(readers, msgchan, &counter)

	for {
		select {
		case msg, ok := <-msgchan:
			now := time.Now()

			if !ok {
				logf("reached EOF, waiting for all write operations to complete...")
				limits.Force = true
				flushAll(dests, store, limits, now, join)
				join.Wait()
				logf("done")
				return
			}

			group, stream := store.Add(msg, now)
			flush(dests, group, stream, limits, now, join)

		case <-expchan:
			logf("timer pulse, flushing streams that haven't been in a while and clearing expired cache...")
			now := time.Now()
			flushAll(dests, store, limits, now, join)
			store.RemoveExpired(cacheTimeout, now)

		case <-sigchan:
			logf("got interrupt signal, closing message reader...")
			stopReaders(readers)
		}
	}
}

func openSources(sources []ecslogs.Source) (readers []ecslogs.Reader, err error) {
	readers = make([]ecslogs.Reader, 0, len(sources))

	for _, source := range sources {
		if r, e := source.Open(); e != nil {
			errorf("failed to open log source (%s)", err)
		} else {
			readers = append(readers, r)
		}
	}

	if len(readers) == 0 {
		err = fmt.Errorf("no sources to read from")
	}

	return
}

func startReaders(readers []ecslogs.Reader, msgchan chan<- ecslogs.Message, counter *int32) {
	hostname, _ := os.Hostname()

	for _, reader := range readers {
		go read(reader, msgchan, counter, hostname)
	}
}

func stopReaders(readers []ecslogs.Reader) {
	for _, reader := range readers {
		reader.Close()
	}
}

func term(c chan<- ecslogs.Message, counter *int32) {
	if atomic.AddInt32(counter, -1) == 0 {
		close(c)
	}
}

func read(r ecslogs.Reader, c chan<- ecslogs.Message, counter *int32, hostname string) {
	defer term(c, counter)
	for {
		var msg ecslogs.Message
		var err error

		if msg, err = r.ReadMessage(); err != nil {
			if err == io.EOF {
				break
			}
			errorf("the message reader failed (%s)", err)
			continue
		}

		if len(msg.Host) == 0 {
			msg.Host = hostname
		}

		if msg.Time == (time.Time{}) {
			msg.Time = time.Now()
		}

		c <- msg
	}
}

func write(dest ecslogs.Destination, group string, stream string, batch []ecslogs.Message, join *sync.WaitGroup) {
	defer join.Done()

	var writer ecslogs.Writer
	var err error

	if writer, err = dest.Open(group, stream); err != nil {
		errorf("dropping message batch of %d messages to %s::%s (%s)", len(batch), group, stream, err)
		return
	}
	defer writer.Close()

	if err = writer.WriteMessageBatch(batch); err != nil {
		errorf("dropping message batch of %d messages to %s::%s (%s)", len(batch), group, stream, err)
		return
	}
}

func flush(dests []ecslogs.Destination, group *ecslogs.Group, stream *ecslogs.Stream, limits ecslogs.StreamLimits, now time.Time, join *sync.WaitGroup) {
	if batch := stream.Flush(limits, now); len(batch) != 0 {
		logf("flushing %d messages to %s::%s", len(batch), group.Name(), stream.Name())
		for _, dest := range dests {
			join.Add(1)
			go write(dest, group.Name(), stream.Name(), batch, join)
		}
	}
}

func flushAll(dests []ecslogs.Destination, store *ecslogs.Store, limits ecslogs.StreamLimits, now time.Time, join *sync.WaitGroup) {
	store.ForEach(func(group *ecslogs.Group) {
		group.ForEach(func(stream *ecslogs.Stream) {
			flush(dests, group, stream, limits, now, join)
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
