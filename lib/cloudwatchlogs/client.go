package cloudwatchlogs

import (
	"errors"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/segmentio/ecs-logs/lib"
)

type client struct {
	cmtx   sync.Mutex
	client *cloudwatchlogs.CloudWatchLogs

	wmtx    sync.Mutex
	writers map[string]*writer
}

func newClient() *client {
	return &client{
		writers: make(map[string]*writer, 100),
	}
}

func (c *client) Open(group string, stream string) (w ecslogs.Writer, err error) {
	var client *cloudwatchlogs.CloudWatchLogs
	var token string
	var writer = c.get(group, stream)

	w = writer

	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	if len(writer.token) != 0 {
		// The writer already has a token, this means the log group and streams
		// have been created for that writer already.
		return
	}

	if client, err = c.getAwsClient(); err != nil {
		return
	}

	if token, err = createGroupAndStream(client, group, stream); err != nil {
		// Creating the log group or stream failed, this writer cannot be used.
		delete(c.writers, joinGroupStream(group, stream))
		return
	}

	writer.token = token
	return
}

func (c *client) Close(group string, stream string) {
	c.remove(group, stream)
}

func (c *client) get(group string, stream string) (w *writer) {
	key := joinGroupStream(group, stream)
	c.wmtx.Lock()

	if w = c.writers[key]; w == nil {
		w = &writer{
			group:  group,
			stream: stream,
			parent: c,
		}
		c.writers[key] = w
	}

	c.wmtx.Unlock()
	return
}

func (c *client) remove(group string, stream string) {
	key := joinGroupStream(group, stream)
	c.wmtx.Lock()
	delete(c.writers, key)
	c.wmtx.Unlock()
}

func (c *client) getAwsClient() (client *cloudwatchlogs.CloudWatchLogs, err error) {
	c.cmtx.Lock()
	defer c.cmtx.Unlock()

	if client = c.client; client == nil {
		if client, err = openAwsClient(); err != nil {
			return
		}
		c.client = client
	}

	return
}

func openAwsClient() (client *cloudwatchlogs.CloudWatchLogs, err error) {
	var region string

	if region, err = getAwsRegion(); err != nil {
		return
	}

	client = cloudwatchlogs.New(session.New(&aws.Config{
		Region: aws.String(region),
	}))
	return
}

func createGroupAndStream(client *cloudwatchlogs.CloudWatchLogs, group string, stream string) (token string, err error) {
	var result *cloudwatchlogs.DescribeLogStreamsOutput

	// Ignore failures on group and stream creation, describing the stream will
	// fail later if the group doesn't exist. That way the group creation is
	// idempotent.
	client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(group),
	})
	client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
	})

	if result, err = client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		Limit:               aws.Int64(1),
		LogGroupName:        aws.String(group),
		LogStreamNamePrefix: aws.String(stream),
	}); err != nil {
		// The AWS Go SDK doesn't export error types, this is the best hack I
		// cloud find to check for this specific error type.
		//
		// The documentation says that we can only make 5 calls per second to
		// this endpoint, but we need the sequence token in order to send events
		// to streams that already exist.
		//
		// If we fail to fetch the stream description we still move on without
		// a token and let the retry logic around PutLogEvents attempt to handle
		// the issue.
		if strings.HasPrefix(err.Error(), "ThrottlingException:") {
			err = nil
		}
		return
	}

	if len(result.LogStreams) == 0 {
		err = errDescribeLogStream
		return
	}

	token = aws.StringValue(result.LogStreams[0].UploadSequenceToken)
	return
}

func joinGroupStream(group string, stream string) string {
	return group + "::" + stream
}

var (
	errDescribeLogStream = errors.New("getting the log stream description failed")
)
