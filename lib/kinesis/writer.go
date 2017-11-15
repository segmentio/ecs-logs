package kinesis

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/segmentio/ecs-logs/lib"
)

var (
	client *kinesis.Kinesis
	once   sync.Once
)

type writer struct {
	group  string
	stream string
}

func (w *writer) WriteMessage(m lib.Message) error {
	req := kinesis.PutRecordInput{
		Data:         m.Bytes(),
		PartitionKey: aws.String(fmt.Sprintf("%s:%s", w.group, w.stream)),
		StreamName:   aws.String("logs"),
	}
	if _, err := client.PutRecord(&req); err != nil {
		return err
	}
	return nil
}

func (w *writer) WriteMessageBatch(m lib.MessageBatch) error {
	records := make([]*kinesis.PutRecordsRequestEntry, m.Len())
	key := fmt.Sprintf("%s:%s", w.group, w.stream)
	for i, s := range m {
		records[i] = &kinesis.PutRecordsRequestEntry{
			Data:         s.Bytes(),
			PartitionKey: aws.String(key),
		}
	}
	req := kinesis.PutRecordsInput{
		Records:    records,
		StreamName: aws.String("logs"),
	}
	if _, err := client.PutRecords(&req); err != nil {
		return err
	}
	return nil
}

func (w *writer) Close() error {
	return nil
}

func NewWriter(group, stream string) (lib.Writer, error) {
	var err error
	once.Do(func() {
		var s *session.Session
		s, err = session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("KINESIS_REGION")),
		})
		if err != nil {
			return
		}
		client = kinesis.New(s)
		err = createStream()
	})
	if err != nil {
		return nil, err
	}

	w := writer{
		group:  group,
		stream: stream,
	}
	return &w, nil
}

type multiError []error

func (m multiError) Error() string {
	s := "error creating stream:\n"
	for _, err := range m {
		s += fmt.Sprintf("\t%v\n", err)
	}
	return s
}

// client.CreateStream doesn't block until the stream
// is created, so we need to do a little dance in order
// to be sure it's available.
func createStream() error {
	req := kinesis.CreateStreamInput{
		ShardCount: aws.Int64(1),
		StreamName: aws.String("logs"),
	}
	var errs multiError
	if _, err := client.CreateStream(&req); err != nil {
		errs = append(errs, err)
	}

	for i := 0; i < 3; i++ {
		req := kinesis.DescribeStreamInput{
			StreamName: aws.String("logs"),
		}
		if _, err := client.DescribeStream(&req); err == nil {
			errs = nil
			break
		} else {
			errs = append(errs, err)
		}

		time.Sleep(500 * time.Millisecond)
	}
	if errs != nil {
		return errs
	}
	return nil
}
