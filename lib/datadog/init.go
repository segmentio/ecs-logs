package datadog

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("datadog", lib.DestinationFunc(NewWriter))
}
