package ecslogs

import (
	"fmt"
	"time"
)

type Stream struct {
	name      string
	bytes     int
	messages  []Message
	createdOn time.Time
	updatedOn time.Time
	flushedOn time.Time
}

type StreamLimits struct {
	MaxCount int
	MaxBytes int
	MaxTime  time.Duration
}

func NewStream(name string, now time.Time) *Stream {
	return &Stream{
		name:      name,
		messages:  make([]Message, 0, 1000),
		createdOn: now,
		updatedOn: now,
	}
}

func (stream *Stream) String() string {
	return fmt.Sprintf("stream { name = %#v }", stream.Name())
}

func (stream *Stream) Name() string {
	return stream.name
}

func (stream *Stream) Add(msg Message, now time.Time) {
	stream.bytes += len(msg.Content)
	stream.messages = append(stream.messages, msg)
	stream.updatedOn = now
}

func (stream *Stream) HasExpired(timeout time.Duration, now time.Time) bool {
	return len(stream.messages) == 0 && now.Sub(stream.updatedOn) >= timeout
}

func (stream *Stream) Flush(limits StreamLimits, now time.Time) []Message {
	if stream.bytes >= limits.MaxBytes {
		return stream.flushDueToBytesLimit(limits.MaxBytes, now)
	}

	if len(stream.messages) >= limits.MaxCount {
		return stream.flushDueToCountLimit(limits.MaxCount, now)
	}

	if now.Sub(stream.flushedOn) >= limits.MaxTime {
		return stream.flushDueToTimeLimit(now)
	}

	return nil
}

func (stream *Stream) flushDueToBytesLimit(maxBytes int, now time.Time) []Message {
	count := 0

	if stream.bytes <= maxBytes {
		count = len(stream.messages)
	} else {
		bytes := 0

		for _, msg := range stream.messages {
			if (bytes + len(msg.Content)) > maxBytes {
				break
			}
			bytes += len(msg.Content)
			count += 1
		}
	}

	return stream.flush(count, now)
}

func (stream *Stream) flushDueToCountLimit(maxCount int, now time.Time) []Message {
	return stream.flush(maxCount, now)
}

func (stream *Stream) flushDueToTimeLimit(now time.Time) []Message {
	return stream.flush(len(stream.messages), now)
}

func (stream *Stream) flush(count int, now time.Time) (msglist []Message) {
	msglist, stream.messages = splitMessageListHead(stream.messages, count)
	stream.flushedOn = now
	return
}

func splitMessageListHead(list []Message, count int) (head []Message, tail []Message) {
	head = make([]Message, count)
	tail = make([]Message, len(list)-count, cap(list))
	copy(head, list[:count])
	copy(tail, list[count:])
	return
}
