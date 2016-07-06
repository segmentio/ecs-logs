package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/multi"
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
	var hostname string
	var level = lib.LogLevel(log.InfoLevel)
	var maxBytes int
	var maxCount int
	var flushTimeout time.Duration
	var cacheTimeout time.Duration

	hostname, _ = os.Hostname()
	log.SetLevel(log.DebugLevel)
	log.SetHandler(cli.New(os.Stderr))

	flag.StringVar(&src, "src", "stdin", "A comma separated list of log sources from which messages will be read ["+strings.Join(lib.SourcesAvailable(), ", ")+"]")
	flag.StringVar(&dst, "dst", "stdout", "A comma separated list of log destinations to which messages will be written ["+strings.Join(lib.DestinationsAvailable(), ", ")+"]")
	flag.StringVar(&hostname, "hostname", hostname, "The hostname advertised by ecs-logs")
	flag.Var(&level, "log-level", "The minimum level of log messages shown by ecs-logs")
	flag.IntVar(&maxBytes, "max-batch-bytes", 1000000, "The maximum size in bytes of a message batch")
	flag.IntVar(&maxCount, "max-batch-size", 10000, "The maximum number of messages in a batch")
	flag.DurationVar(&flushTimeout, "flush-timeout", 5*time.Second, "How often messages will be flushed")
	flag.DurationVar(&cacheTimeout, "cache-timeout", 5*time.Minute, "How to wait before clearing unused internal cache")
	flag.Parse()

	var store = lib.NewStore()
	var sources []source
	var readers []reader
	var dests []destination

	if len(hostname) == 0 {
		log.Fatal("no hostname configured")
	}

	if sources = getSources(strings.Split(src, ",")); len(sources) == 0 {
		log.Fatal("no or invalid log sources")
	}

	if dests = getDestinations(strings.Split(dst, ",")); len(dests) == 0 {
		log.Fatal("no or invalid log destinations")
	}

	if readers, err = openSources(sources); err != nil {
		log.WithError(err).Fatal("failed to open log sources readers")
	}

	join := &sync.WaitGroup{}

	limits := lib.StreamLimits{
		MaxCount: maxCount,
		MaxBytes: maxBytes,
		MaxTime:  flushTimeout,
	}

	logger := &lib.LogHandler{
		Group:    "ecs-logs",
		Stream:   hostname,
		Hostname: hostname,
		Queue:    lib.NewMessageQueue(),
	}
	log.SetLevel(log.Level(level))
	log.SetHandler(multi.New(cli.New(os.Stderr), logger))

	expchan := time.Tick(flushTimeout / 2)
	msgchan := make(chan lib.Message, len(readers))
	sigchan := make(chan os.Signal, 1)
	counter := int32(len(readers))
	startReaders(readers, msgchan, &counter, hostname)
	setupSignals(sigchan)

	for {
		select {
		case msg, ok := <-msgchan:
			now := time.Now()

			if !ok {
				log.Info("waiting for all write operations to complete")
				limits.Force = true
				flushAll(dests, store, limits, now, join)
				flushQueue(dests, store, logger.Queue, limits, now, join)
				join.Wait()
				return
			}

			_, stream := store.Add(msg, now)
			flush(dests, stream, limits, now, join)

		case <-logger.Queue.C:
			now := time.Now()
			flushQueue(dests, store, logger.Queue, limits, now, join)

		case <-expchan:
			now := time.Now()
			flushAll(dests, store, limits, now, join)
			removeExpired(dests, store, cacheTimeout, now)

		case sig := <-sigchan:
			log.WithFields(log.Fields{"signal": sig.String()}).Info("closing message readers")
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
		log.WithField("source", names[i]).Info("source enabled")
	}
	return
}

func getDestinations(names []string) (destinations []destination) {
	for i, dst := range lib.GetDestinations(names...) {
		destinations = append(destinations, destination{
			Destination: dst,
			name:        names[i],
		})
		log.WithField("destination", names[i]).Info("destination enabled")
	}
	return
}

func openSources(sources []source) (readers []reader, err error) {
	readers = make([]reader, 0, len(sources))

	for _, source := range sources {
		if r, e := source.Open(); e != nil {
			log.WithFields(log.Fields{
				"source": source.name,
				"error":  e,
			}).Error("failed to open log source")
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

func setupSignals(sigchan chan<- os.Signal) {
	signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
}

func startReaders(readers []reader, msgchan chan<- lib.Message, counter *int32, hostname string) {
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
				log.WithFields(log.Fields{
					"reader": r.name,
				}).Info("the message reader was closed")
			} else {
				log.WithFields(log.Fields{
					"reader": r.name,
					"error":  err,
				}).Error("the message reader failed")
			}
			return
		}

		if len(msg.Group) == 0 {
			log.WithFields(log.Fields{
				"reader":  r.name,
				"missing": "group",
			}).Warn("dropping message because the a required field wasn't set")
			continue
		}

		if len(msg.Stream) == 0 {
			log.WithFields(log.Fields{
				"reader":  r.name,
				"missing": "stream",
			}).Warn("dropping message because the a required field wasn't set")
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

func write(dest destination, group string, stream string, batch lib.MessageBatch, join *sync.WaitGroup) {
	defer join.Done()

	var writer lib.Writer
	var err error

	if writer, err = dest.Open(group, stream); err != nil {
		logDropBatch(dest.name, group, stream, err, batch)
		return
	}
	defer writer.Close()

	if err = writer.WriteMessageBatch(batch); err != nil {
		logDropBatch(dest.name, group, stream, err, batch)
		return
	}
}

func flush(dests []destination, stream *lib.Stream, limits lib.StreamLimits, now time.Time, join *sync.WaitGroup) {
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

		log.WithFields(log.Fields{
			"group":  stream.Group(),
			"stream": stream.Name(),
			"count":  len(batch),
			"reason": reason,
		}).Info("flushing message batch")

		for _, dest := range dests {
			join.Add(1)
			go write(dest, stream.Group(), stream.Name(), batch, join)
		}
	}
}

func flushAll(dests []destination, store *lib.Store, limits lib.StreamLimits, now time.Time, join *sync.WaitGroup) {
	store.ForEach(func(group *lib.Group) {
		group.ForEach(func(stream *lib.Stream) {
			flush(dests, stream, limits, now, join)
		})
	})
}

func flushQueue(dests []destination, store *lib.Store, queue *lib.MessageQueue, limits lib.StreamLimits, now time.Time, join *sync.WaitGroup) {
	streams := make(map[string]*lib.Stream)

	for _, msg := range queue.Flush() {
		_, stream := store.Add(msg, now)
		key := stream.Group() + ":" + stream.Name()

		if streams[key] == nil {
			streams[key] = stream
		}
	}

	for _, stream := range streams {
		flush(dests, stream, limits, now, join)
	}
}

func removeExpired(dests []destination, store *lib.Store, cacheTimeout time.Duration, now time.Time) {
	for _, stream := range store.RemoveExpired(cacheTimeout, now) {
		for _, dest := range dests {
			dest.Close(stream.Group(), stream.Name())
		}
		log.WithFields(log.Fields{
			"group":  stream.Group(),
			"stream": stream.Name(),
		}).Info("removed expired stream")
	}
}

func logDropBatch(dest string, group string, stream string, err error, batch lib.MessageBatch) {
	log.WithFields(log.Fields{
		"group":       group,
		"stream":      stream,
		"destination": dest,
		"error":       err,
		"count":       len(batch),
	}).Error("dropping message batch")

	for _, msg := range batch {
		log.WithFields(log.Fields{
			"group":  msg.Group,
			"stream": msg.Stream,
			"event":  msg.Event,
		}).Debug("dropped")
	}
}
