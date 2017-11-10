package cloudwatchlogs

import (
	"errors"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/segmentio/ecs-logs/lib"
)

type writer struct {
	mutex  sync.Mutex
	group  string
	stream string
	token  string
	parent *client
}

func (w *writer) Close() error {
	return nil
}

func (w *writer) WriteMessage(msg lib.Message) error {
	return w.WriteMessageBatch(lib.MessageBatch{msg})
}

func (w *writer) WriteMessageBatch(batch lib.MessageBatch) (err error) {
	if len(batch) == 0 {
		return
	}

	var token *string
	var result *cloudwatchlogs.PutLogEventsOutput
	var events = make([]*cloudwatchlogs.InputLogEvent, len(batch))

	for i, msg := range batch {
		events[i] = &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(msg.GoodEggsEventString()),
			Timestamp: aws.Int64(aws.TimeUnixMilli(msg.Event.Time)),
		}
	}

	// Because of the logic imposed by the AWS API we can only submit one upload
	// request per log stream at a time due to the sequence token being unique
	// and usable only once.
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.parent == nil {
		// Another goroutine has invalidated this writer, giving up.
		err = errInvalidWriter
		return
	}

	if len(w.token) != 0 {
		token = aws.String(w.token)
	}

	for attempt := 1; true; attempt++ {
		if result, err = w.parent.client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
			LogEvents:     events,
			LogGroupName:  aws.String(w.group),
			LogStreamName: aws.String(w.stream),
			SequenceToken: token,
		}); err == nil {
			break
		}

		// The AWS Go SDK doesn't expose the error type but does return the
		// token in the error message so we attempt to extract it from there
		// and let the retry logic resubmit the event batch.
		//
		// See: https://forums.aws.amazon.com/message.jspa?messageID=676912
		if token = parseInvalidSequenceTokenException(err); attempt < 3 && token != nil {
			err = nil
			continue
		}

		// The documentation says we have to provide the sequence token when
		// uploading events to CloudWatchLogs, if an error is returned here
		// it's likely the token we have is either invalid or something worse
		// happened.
		// We remove the writer from it's parent client so a new writer will
		// be created.
		w.parent.remove(w.group, w.stream)
		w.parent = nil
		return
	}

	w.token = aws.StringValue(result.NextSequenceToken)
	return
}

func parseInvalidSequenceTokenException(err error) (token *string) {
	msg := err.Error()

	if !strings.HasPrefix(msg, "InvalidSequenceTokenException:") {
		return
	}

	if lines := strings.Split(msg, "\n"); len(lines) != 0 {
		msg = lines[0]
	}

	parts := strings.Split(msg, ":")

	if len(parts) < 3 {
		return
	}

	s := strings.TrimSpace(parts[2])
	token = &s
	return
}

var (
	errInvalidWriter = errors.New("the writer was invalidated by another goroutine")
)
