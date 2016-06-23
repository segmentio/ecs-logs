package syslog

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterDestination("syslog", ecslogs.DestinationFunc(NewMessageBatchWriter))
}
