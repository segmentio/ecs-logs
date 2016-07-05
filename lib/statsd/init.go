package statsd

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterDestination("statsd", ecslogs.DestinationFunc(NewWriter))
}
