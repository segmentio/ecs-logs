package syslog

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterDestination("syslog", lib.DestinationFunc(NewWriter))
}
