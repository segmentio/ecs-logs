package cloudwatchlogs

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterDestination("cloudwatchlogs", newClient())
}
