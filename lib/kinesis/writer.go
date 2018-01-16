package kinesis

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/segmentio/ecs-logs/lib"
)

// Safe for concurrent use
var client *kinesis.Kinesis

type writer struct {
	group  string // log group, i.e. service name.
	stream string // log stream, i.e. host name.
}

func (w *writer) WriteMessage(m lib.Message) error {
	req := kinesis.PutRecordInput{
		Data:         m.Bytes(),
		PartitionKey: aws.String(w.stream),
		StreamName:   aws.String(w.group),
	}
	_, err := client.PutRecord(&req)
	return err
}

func (w *writer) WriteMessageBatch(m lib.MessageBatch) error {
	records := make([]*kinesis.PutRecordsRequestEntry, m.Len())
	for i, s := range m {
		records[i] = &kinesis.PutRecordsRequestEntry{
			Data:         s.Bytes(),
			PartitionKey: aws.String(w.stream),
		}
	}
	req := kinesis.PutRecordsInput{
		Records:    records,
		StreamName: aws.String(w.group),
	}
	_, err := client.PutRecords(&req)
	return err
}

func (w *writer) Close() error {
	return nil
}

func NewWriter(group, stream string) (lib.Writer, error) {
	if client == nil {
		var s *session.Session
		s, err := session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("KINESIS_REGION")),
		})
		if err != nil {
			return nil, err
		}
		client = kinesis.New(s)
	}

	if err := checkStream(group); err != nil {
		return nil, err
	}

	w := writer{
		group:  group,
		stream: stream,
	}
	return &w, nil
}

func checkStream(group string) error {
	req := kinesis.DescribeStreamInput{
		StreamName: aws.String(group),
	}
	_, err := client.DescribeStream(&req)
	return err
}
