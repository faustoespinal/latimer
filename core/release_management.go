package core

import (
	"latimer/kube"
	"time"

	"github.com/sirupsen/logrus"
)

// WaitForRelease pauses for up to 'timeout' seconds waiting for the specified release to be fully installed
func WaitForRelease(sc *SystemContext, installable Installable, timeout time.Duration) bool {
	start := time.Now()
	for installable.Status(sc) != kube.Ready {
		time.Sleep(2 * time.Second)
		end := time.Now()
		elapsed := end.Sub(start)
		if elapsed > timeout {
			return false
		}
		logrus.Debugf("       Waiting for release %v Elapsed=%v\n", installable.GetID, elapsed)
	}
	return true
}
