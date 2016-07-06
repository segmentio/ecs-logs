package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/ecs-logs/lib"

	_ "github.com/segmentio/ecs-logs/lib/cloudwatchlogs"
	_ "github.com/segmentio/ecs-logs/lib/datadog"
	_ "github.com/segmentio/ecs-logs/lib/loggly"
	_ "github.com/segmentio/ecs-logs/lib/statsd"
	_ "github.com/segmentio/ecs-logs/lib/syslog"
)

type source struct {
	lib.Source
	name string
}

type destination struct {
	lib.Destination
	name string
}

type reader struct {
	lib.Reader
	name string
}

func main() {
	var err error
	var src string
	var dst string
	var maxBytes int
	var maxCount int
	var flushTimeout time.Duration
	var cacheTimeout time.Duration

	flag.StringVar(&src, "src", "stdin", "A comma separated list of log sources from which messages will be read ["+strings.Join(lib.SourcesAvailable(), ", ")+"]")
	flag.StringVar(&dst, "dst", "stdout", "A comma separated list of log destinations to which messages will be written ["+strings.Join(lib.DestinationsAvailable(), ", ")+"]")
	flag.IntVar(&maxBytes, "max-batch-bytes", 1000000, "The maximum size in bytes of a message batch")
	flag.IntVar(&maxCount, "max-batch-size", 10000, "The maximum number of messages in a batch")
	flag.DurationVar(&flushTimeout, "flush-timeout", 5*time.Second, "How often messages will be flushed")
	flag.DurationVar(&cacheTimeout, "cache-timeout", 5*time.Minute, "How to wait before clearing unused internal cache")
	flag.Parse()

	var store = lib.NewStore()
	var sources []source
	var readers []reader
	var dests []destination

	if sources = getSources(strings.Split(src, ",")); len(sources) == 0 {
		fatalf("no or invalid log sources")
	}

	if dests = getDestinations(strings.Split(dst, ",")); len(dests) == 0 {
		fatalf("no or invalid log destinations")
	}

	if readers, err = openSources(sources); err != nil {
		fatalf("failed to open sources (%s)", err)
	}

	join := &sync.WaitGroup{}

	limits := lib.StreamLimits{
		MaxCount: maxCount,
		MaxBytes: maxBytes,
		MaxTime:  flushTimeout,
	}

	expchan := time.Tick(flushTimeout / 2)
	msgchan := make(chan lib.Message, len(readers))
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
			now := time.Now()
			flushAll(dests, store, limits, now, join)
			removeExpired(dests, store, cacheTimeout, now)

		case <-sigchan:
			logf("got interrupt signal, closing message reader...")
			stopReaders(readers)
		}
	}
}

func getSources(names []string) (sources []source) {
	for i, src := range lib.GetSources(names...) {
		sources = append(sources, source{
			Source: src,
			name:   names[i],
		})
	}
	return
}

func getDestinations(names []string) (destinations []destination) {
	for i, dst := range lib.GetDestinations(names...) {
		destinations = append(destinations, destination{
			Destination: dst,
			name:        names[i],
		})
	}
	return
}

func openSources(sources []source) (readers []reader, err error) {
	readers = make([]reader, 0, len(sources))

	for _, source := range sources {
		if r, e := source.Open(); e != nil {
			errorf("failed to open log source (%s: %s)", source.name, e)
		} else {
			readers = append(readers, reader{
				Reader: r,
				name:   source.name,
			})
		}
	}

	if len(readers) == 0 {
		err = fmt.Errorf("no sources to read from")
	}

	return
}

func startReaders(readers []reader, msgchan chan<- lib.Message, counter *int32) {
	hostname, _ := os.Hostname()

	for _, reader := range readers {
		go read(reader, msgchan, counter, hostname)
	}
}

func stopReaders(readers []reader) {
	for _, reader := range readers {
		reader.Close()
	}
}

func term(c chan<- lib.Message, counter *int32) {
	if atomic.AddInt32(counter, -1) == 0 {
		close(c)
	}
}

func read(r reader, c chan<- lib.Message, counter *int32, hostname string) {
	defer term(c, counter)
	for {
		var msg lib.Message
		var err error

		if msg, err = r.ReadMessage(); err != nil {
			if err == io.EOF {
				break
			}
			errorf("the message reader failed (%s: %s)", r.name, err)
			continue
		}

		if len(msg.Group) == 0 {
			errorf("dropping %s message because the group property wasn't set", r.name)
			errorf("- %s", msg.Event)
			continue
		}

		if len(msg.Stream) == 0 {
			errorf("dropping %s message because the stream property wasn't set (%s)", r.name, msg.Group)
			errorf("- %s", msg.Event)
			continue
		}

		if len(msg.Event.Info.Host) == 0 {
			msg.Event.Info.Host = hostname
		}

		if msg.Event.Time == (time.Time{}) {
			msg.Event.Time = time.Now()
		}

		if msg.Event.Data == nil {
			msg.Event.Data = ecslogs.EventData{}
		}

		c <- msg
	}
}

func write(dest destination, group string, stream string, batch []lib.Message, join *sync.WaitGroup) {
	defer join.Done()

	var writer lib.Writer
	var err error

	if writer, err = dest.Open(group, stream); err != nil {
		errorBatch(dest.name, group, stream, err, batch)
		return
	}
	defer writer.Close()

	if err = writer.WriteMessageBatch(batch); err != nil {
		errorBatch(dest.name, group, stream, err, batch)
		return
	}
}

func flush(dests []destination, group *lib.Group, stream *lib.Stream, limits lib.StreamLimits, now time.Time, join *sync.WaitGroup) {
	for {
		batch, reason := stream.Flush(limits, now)

		if len(batch) == 0 {
			break
		}

		// Ensure all messages in the batch are sorted. Checking if the batch is
		// sorted is an optimization since in most cases the batch will be sorted
		// because we're reading events that are generated live (checking for a
		// sorted list is O(N) vs O(N*log(N)) for sorting it).
		// There are cases where some log entries do appear unordered and this is
		// causing issues with CloudWatchLogs.
		if !sort.IsSorted(batch) {
			sort.Stable(batch)
		}

		logf("flushing %d messages to %s::%s (%s)", len(batch), group.Name(), stream.Name(), reason)

		for _, dest := range dests {
			join.Add(1)
			go write(dest, group.Name(), stream.Name(), batch, join)
		}
	}
}

func flushAll(dests []destination, store *lib.Store, limits lib.StreamLimits, now time.Time, join *sync.WaitGroup) {
	store.ForEach(func(group *lib.Group) {
		group.ForEach(func(stream *lib.Stream) {
			flush(dests, group, stream, limits, now, join)
		})
	})
}

func removeExpired(dests []destination, store *lib.Store, cacheTimeout time.Duration, now time.Time) {
	for _, stream := range store.RemoveExpired(cacheTimeout, now) {
		for _, dest := range dests {
			logf("removed expired stream %s::%s (%s)", stream.Group(), stream.Name(), dest.name)
			dest.Close(stream.Group(), stream.Name())
		}
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func errorBatch(dest string, group string, stream string, err error, batch []lib.Message) {
	errorf("dropping message batch of %d messages to %s::%s (%s: %s)", len(batch), group, stream, dest, err)
	for _, msg := range batch {
		errorf("- %s", msg.Event)
	}
}
