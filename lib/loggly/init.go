package loggly

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterDestination("loggly", ecslogs.DestinationFunc(NewWriter))
}
