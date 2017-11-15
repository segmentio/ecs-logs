package kinesis

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("kinesis", lib.DestinationFunc(NewWriter))
}
