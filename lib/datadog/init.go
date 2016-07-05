package datadog

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterDestination("datadog", ecslogs.DestinationFunc(NewWriter))
}
