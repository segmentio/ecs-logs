package ecslogs

import (
	"os"
	"sort"
	"sync"
)

type Source interface {
	Open() (MessageReadCloser, error)
}

type SourceFunc func() (MessageReadCloser, error)

func (f SourceFunc) Open() (MessageReadCloser, error) {
	return f()
}

func RegisterSource(name string, source Source) {
	srcmtx.Lock()
	srcmap[name] = source
	srcmtx.Unlock()
}

func DeregisterSource(name string) {
	srcmtx.Lock()
	delete(srcmap, name)
	srcmtx.Unlock()
}

func GetSource(name string) (source Source) {
	srcmtx.RLock()
	source = srcmap[name]
	srcmtx.RUnlock()
	return
}

func SourcesAvailable() (sources []string) {
	srcmtx.RLock()
	sources = make([]string, 0, len(srcmap))

	for name := range srcmap {
		sources = append(sources, name)
	}

	srcmtx.RUnlock()
	sort.Strings(sources)
	return
}

var (
	srcmtx sync.RWMutex
	srcmap = map[string]Source{
		"stdin": SourceFunc(func() (MessageReadCloser, error) {
			return NewMessageDecoder(os.Stdin), nil
		}),
	}
)
