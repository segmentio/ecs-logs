package syslog

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterSource("syslog", ecslogs.SourceFunc(NewMessageReader))
	ecslogs.RegisterDestination("syslog", ecslogs.DestinationFunc(NewMessageBatchWriter))
}
