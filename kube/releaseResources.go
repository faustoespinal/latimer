package kube

import (
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	jobsv1 "k8s.io/api/batch/v1"
)

// InstallStatus models the different status codes for an installable
type InstallStatus int

const (
	// NotReady means the Installable is in process of installation/uninstallation but not ready
	NotReady InstallStatus = iota
	// Ready means Installable is fully installed and operational
	Ready
	// NotInstalled means that nothing has been installed for the installable (no resource artifacts exist)
	NotInstalled
)

// ReleaseResources collects core resources under a release
type ReleaseResources struct {
	// The name of the release
	ReleaseName string

	Deployments  []appsv1.Deployment
	StatefulSets []appsv1.StatefulSet
	DaemonSets   []appsv1.DaemonSet
	Jobs         []jobsv1.Job
}

// NewReleaseResources creates a new instance of ReleaseResources
func NewReleaseResources(releaseName string) *ReleaseResources {
	relResources := new(ReleaseResources)

	relResources.ReleaseName = releaseName
	relResources.Deployments = make([]appsv1.Deployment, 0)
	relResources.StatefulSets = make([]appsv1.StatefulSet, 0)
	relResources.DaemonSets = make([]appsv1.DaemonSet, 0)
	relResources.Jobs = make([]jobsv1.Job, 0)
	return relResources
}

// ReleaseStatus indicates whether the state of the release ready or not
func (rr *ReleaseResources) ReleaseStatus() InstallStatus {
	totalResources := 0
	for _, depl := range rr.Deployments {
		totalResources++
		logrus.Debugf("           DEPL: name=%v readyReplicas=%v replicas=%v\n", depl.Name, depl.Status.ReadyReplicas, *depl.Spec.Replicas)
		if depl.Status.ReadyReplicas < *depl.Spec.Replicas {
			return NotReady
		}
	}
	for _, ss := range rr.StatefulSets {
		totalResources++
		logrus.Debugf("           SS  : name=%v readyReplicas=%v replicas=%v\n", ss.Name, ss.Status.ReadyReplicas, *ss.Spec.Replicas)
		if ss.Status.ReadyReplicas < *ss.Spec.Replicas {
			return NotReady
		}
	}
	for _, dd := range rr.DaemonSets {
		totalResources++
		if dd.Status.NumberUnavailable > 0 {
			return NotReady
		}
	}
	for _, job := range rr.Jobs {
		totalResources++
		logrus.Debugf("           JOB : name=%v NumSucceeded=%v Completions=%v\n", job.Name, job.Status.Succeeded, *job.Spec.Completions)
		if job.Status.Succeeded < *job.Spec.Completions {
			return NotReady
		}
	}
	// If there are no runtime resources associated to the release, then it is not installed
	if totalResources == 0 {
		return NotInstalled
	}
	return Ready
}
