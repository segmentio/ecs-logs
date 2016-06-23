package cloudwatchlogs

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/segmentio/ecs-logs/lib"
)

func NewMessageBatchWriter(group string, stream string) (w ecslogs.MessageBatchWriteCloser, err error) {
	var region string

	if region, err = getAwsRegion(); err != nil {
		return
	}

	w = messageWriter{
		group:  group,
		stream: stream,
		client: cloudwatchlogs.New(session.New(&aws.Config{
			Region: aws.String(region),
		})),
	}
	return
}

type messageWriter struct {
	group  string
	stream string
	client *cloudwatchlogs.CloudWatchLogs
}

func (w messageWriter) Close() error {
	return nil
}

func (w messageWriter) WriteMessage(msg ecslogs.Message) error {
	return w.WriteMessageBatch([]ecslogs.Message{msg})
}

func (w messageWriter) WriteMessageBatch(batch []ecslogs.Message) (err error) {
	var stream *cloudwatchlogs.LogStream

	w.ensureCreateGroup()
	w.ensureCreateStream()

	if stream, err = w.fetchStream(); err != nil {
		return
	}

	return w.writeBatch(stream, batch)
}

func (w messageWriter) ensureCreateGroup() {
	w.client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(w.group),
	})
	return
}

func (w messageWriter) ensureCreateStream() {
	w.client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(w.group),
		LogStreamName: aws.String(w.stream),
	})
	return
}

func (w messageWriter) fetchStream() (stream *cloudwatchlogs.LogStream, err error) {
	var streams *cloudwatchlogs.DescribeLogStreamsOutput

	if streams, err = w.client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(w.group),
		LogStreamNamePrefix: aws.String(w.stream),
		Limit:               aws.Int64(1),
	}); err != nil {
		return
	}

	if len(streams.LogStreams) == 0 {
		err = fmt.Errorf("failed to fetch log stream info from cloud watch logs (%s: stream not found)", w.stream)
		return
	}

	stream = streams.LogStreams[0]
	return
}

func (w messageWriter) writeBatch(stream *cloudwatchlogs.LogStream, batch []ecslogs.Message) (err error) {
	var events []*cloudwatchlogs.InputLogEvent

	if events = makeLogEvents(batch); len(events) == 0 {
		return
	}

	_, err = w.client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(w.group),
		LogStreamName: aws.String(w.stream),
		SequenceToken: stream.UploadSequenceToken,
	})
	return
}

func makeLogEvents(batch []ecslogs.Message) (events []*cloudwatchlogs.InputLogEvent) {
	events = make([]*cloudwatchlogs.InputLogEvent, 0, len(batch))

	for _, msg := range batch {
		if len(msg.Content) != 0 {
			msg.Time = time.Time{}
			events = append(events, &cloudwatchlogs.InputLogEvent{
				Message:   aws.String(msg.String()),
				Timestamp: aws.Int64(aws.TimeUnixMilli(msg.Time)),
			})
		}
	}

	return
}
