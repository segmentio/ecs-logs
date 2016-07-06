package lib

import (
	"io"
	"os"
	"sort"
	"sync"
)

type Source interface {
	Open() (Reader, error)
}

type SourceFunc func() (Reader, error)

func (f SourceFunc) Open() (Reader, error) {
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

func GetSources(names ...string) (sources []Source) {
	sources = make([]Source, 0, len(names))

	for _, name := range names {
		if source := GetSource(name); source != nil {
			sources = append(sources, source)
		}
	}

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
		"stdin": SourceFunc(func() (Reader, error) {
			// On some platforms closing stdin doesn't cause pending read
			// operations to abort, resulting in a blocking call that never
			// returns.
			//
			// To work around this limitation we start a goroutine that is
			// responsible for reading from stdin and send the bytes through
			// an in-memory pipe. When the pipe is closed it properly cancels
			// all pending reads which is the behavior we expect.
			//
			// There probably is a small performance cost to adding this extra
			// step but the stdin source shouldn't be used in production
			// environments so it shouldn't be a problem in practice.
			//
			// Note that the goroutine reading from stdin is likely gonna be
			// leaked... This is OK in the ecs-logs use case because only one
			// stdin reader will be instantiated.
			r, w := io.Pipe()
			go io.Copy(w, os.Stdin)
			return NewMessageDecoder(r), nil
		}),
	}
)
