package kube

import (
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	jobsv1 "k8s.io/api/batch/v1"
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

// IsReleaseInstalled indicates whether the state of the release ready or not
func (rr *ReleaseResources) IsReleaseInstalled() bool {
	for _, depl := range rr.Deployments {
		logrus.Debugf("           DEPL: name=%v readyReplicas=%v replicas=%v\n", depl.Name, depl.Status.ReadyReplicas, *depl.Spec.Replicas)
		if depl.Status.ReadyReplicas < *depl.Spec.Replicas {
			return false
		}
	}
	for _, ss := range rr.StatefulSets {
		logrus.Debugf("           SS  : name=%v readyReplicas=%v replicas=%v\n", ss.Name, ss.Status.ReadyReplicas, *ss.Spec.Replicas)
		if ss.Status.ReadyReplicas < *ss.Spec.Replicas {
			return false
		}
	}
	for _, dd := range rr.DaemonSets {
		if dd.Status.NumberUnavailable > 0 {
			return false
		}
	}
	for _, job := range rr.Jobs {
		logrus.Debugf("           JOB : name=%v NumSucceeded=%v Completions=%v\n", job.Name, job.Status.Succeeded, *job.Spec.Completions)
		if job.Status.Succeeded < *job.Spec.Completions {
			return false
		}
	}
	return true
}
