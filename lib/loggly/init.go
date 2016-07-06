package loggly

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("loggly", lib.DestinationFunc(NewWriter))
}
