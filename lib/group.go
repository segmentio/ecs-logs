package ecslogs

import (
	"fmt"
	"time"
)

type Group struct {
	name      string
	streams   map[string]*Stream
	createdOn time.Time
	updatedOn time.Time
}

func NewGroup(name string, now time.Time) *Group {
	return &Group{
		name:      name,
		streams:   make(map[string]*Stream),
		createdOn: now,
		updatedOn: now,
	}
}

func (group *Group) String() string {
	return fmt.Sprintf("group { name = %s }", group.Name())
}

func (group *Group) Name() string {
	return group.name
}

func (group *Group) Add(msg Message, now time.Time) (stream *Stream) {
	if stream = group.streams[msg.Stream]; stream == nil {
		stream = NewStream(group.Name(), msg.Stream, now)
		group.streams[msg.Stream] = stream
	}

	stream.Add(msg, now)
	group.updatedOn = now
	return
}

func (group *Group) HasExpired(timeout time.Duration, now time.Time) bool {
	return len(group.streams) == 0 && now.Sub(group.updatedOn) >= timeout
}

func (group *Group) RemoveExpired(timeout time.Duration, now time.Time) (streams []*Stream) {
	for name, stream := range group.streams {
		if stream.HasExpired(timeout, now) {
			streams = append(streams, stream)
			delete(group.streams, name)
		}
	}
	return
}

func (group *Group) ForEach(f func(*Stream)) {
	for _, stream := range group.streams {
		f(stream)
	}
}
