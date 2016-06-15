package ecslogs

import "time"

type Store struct {
	groups map[string]*Group
}

func NewStore() *Store {
	return &Store{
		groups: make(map[string]*Group, 100),
	}
}

func (store *Store) Add(msg Message, now time.Time) (group *Group, stream *Stream) {
	if group = store.groups[msg.Group]; group == nil {
		group = NewGroup(msg.Group, now)
		store.groups[msg.Group] = group
	}

	stream = group.Add(msg, now)
	return
}

func (store *Store) RemoveExpired(timeout time.Duration, now time.Time) {
	for name, group := range store.groups {
		if group.HasExpired(timeout, now) {
			delete(store.groups, name)
		}
	}
}

func (store *Store) ForEach(f func(*Group)) {
	for _, group := range store.groups {
		f(group)
	}
}
