package cloudwatchlogs

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("cloudwatchlogs", newClient())
}
