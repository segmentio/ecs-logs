// +build linux

package journald

import "github.com/segmentio/ecs-logs/lib"

func init() {
	ecslogs.RegisterSource("journald", ecslogs.SourceFunc(NewReader))
}
