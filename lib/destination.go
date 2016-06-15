package ecslogs

import (
	"os"
	"sort"
	"sync"
)

type Destination interface {
	Open(group string, stream string) (MessageBatchWriteCloser, error)
}

type DestinationFunc func(group string, stream string) (MessageBatchWriteCloser, error)

func (f DestinationFunc) Open(group string, stream string) (MessageBatchWriteCloser, error) {
	return f(group, stream)
}

func RegisterDestination(name string, destination Destination) {
	dstmtx.Lock()
	dstmap[name] = destination
	dstmtx.Unlock()
}

func DeregisterDestination(name string) {
	dstmtx.Lock()
	delete(dstmap, name)
	dstmtx.Unlock()
}

func GetDestination(name string) (destination Destination) {
	dstmtx.RLock()
	destination = dstmap[name]
	dstmtx.RUnlock()
	return
}

func DestinationsAvailable() (destinations []string) {
	dstmtx.RLock()
	destinations = make([]string, 0, len(dstmap))

	for name := range dstmap {
		destinations = append(destinations, name)
	}

	dstmtx.RUnlock()
	sort.Strings(destinations)
	return
}

var (
	dstmtx sync.RWMutex
	dstmap = map[string]Destination{
		"stdout": DestinationFunc(func(_ string, _ string) (MessageBatchWriteCloser, error) {
			return NewMessageEncoder(os.Stdout), nil
		}),
	}
)
