// +build linux

package sysjournald

import "github.com/segmentio/ecs-logs/lib"

func init() {
	lib.RegisterSource("sysjournald", lib.SourceFunc(NewReader))
}
