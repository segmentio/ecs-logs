package statsd

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("statsd", lib.DestinationFunc(NewWriter))
}
