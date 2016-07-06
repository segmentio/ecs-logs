// +build linux

package journald

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterSource("journald", lib.SourceFunc(NewReader))
}
