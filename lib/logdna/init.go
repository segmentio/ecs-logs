package logdna

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("logdna", lib.DestinationFunc(NewWriter))
}
