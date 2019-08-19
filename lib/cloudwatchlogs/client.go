package cloudwatchlogs

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

func (c *client) Open(group string, stream string) (w lib.Writer, err error) {
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
		c.remove(group, stream)
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

	if _, err := client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(group),
	}); err != nil && !isAlreadyExists(err) {
		return "", err
	}

	if _, err = client.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName: aws.String(group),
		RetentionInDays: aws.Int64(180),
	}); err != nil {
                fmt.Println(err.Error())
	}

	_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
	})
	if err == nil {
		// Log stream successfully created.  No token need be provided.
		return "", nil
	} else if !isAlreadyExists(err) {
		return "", err
	}

	if result, err = client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		Limit:               aws.Int64(1),
		LogGroupName:        aws.String(group),
		LogStreamNamePrefix: aws.String(stream),
	}); err != nil {
		if isThrottled(err) {
			// The documentation says that we can only make 5 calls per second to
			// this endpoint, but we need the sequence token in order to send events
			// to streams that already exist.
			//
			// If we fail to fetch the stream description we still move on without
			// a token and let the retry logic around PutLogEvents attempt to handle
			// the issue.
			return "", nil
		}
		return "", err
	}

	// This should be an invariant
	if len(result.LogStreams) == 0 {
		return "", fmt.Errorf("Assertion failure: Log stream %s: %s not found",
			group, stream)
	}

	return aws.StringValue(result.LogStreams[0].UploadSequenceToken), nil
}

func joinGroupStream(group string, stream string) string {
	return group + ":" + stream
}

func isAwsErrorCode(err error, code string) bool {
	if err, ok := err.(awserr.Error); ok && err.Code() == code {
		return true
	}
	return false
}

func isAlreadyExists(err error) bool {
	return isAwsErrorCode(err, "ResourceAlreadyExistsException")
}

func isThrottled(err error) bool {
	return isAwsErrorCode(err, "ThrottlingException")
}
