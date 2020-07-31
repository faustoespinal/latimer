package kube

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// AnnotationReleaseName is the map key to find the release name in the annotation block
	AnnotationReleaseName = "meta.helm.sh/release-name"
	// LabelReleaseName is the map key to find the release name in the label block
	LabelReleaseName = "release"
)

// K8sClient defines a class representing a kubernetes client capable of executing a variety of commands
// against a specified API server.
type K8sClient struct {
	kubeConfigPath string
	kubeConfig     *rest.Config
	clientSet      *kubernetes.Clientset
}

// NewK8sClient creates a new instance of a kubernetes client
func NewK8sClient(kubeConfigPath string) (*K8sClient, error) {
	kubeClient := new(K8sClient)
	config, err := newKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	clientSet, err := newClientSet(config)
	if err != nil {
		return nil, err
	}
	kubeClient.kubeConfigPath = kubeConfigPath
	kubeClient.kubeConfig = config
	kubeClient.clientSet = clientSet
	return kubeClient, nil
}

// newGetKubeConfig loads the kube config settings
func newKubeConfig(kubeConfig string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	return config, err
}

// newClientSet returns the clientset object which to use for issuing k8s API calls.
func newClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

// GetNamespace returns whether the specified namespace name exists
func (k8s *K8sClient) GetNamespace(namespace string) *v1.Namespace {
	ns, err := k8s.clientSet.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil
	}
	return ns
}

// CreateNamespace creates a namespace by the given name.
func (k8s *K8sClient) CreateNamespace(namespace string) (*v1.Namespace, error) {
	ns := v1.Namespace{}
	ns.Name = namespace
	return k8s.clientSet.CoreV1().Namespaces().Create(context.TODO(), &ns, metav1.CreateOptions{})
}

// DeleteNamespace deletes a namespace by the given name.
func (k8s *K8sClient) DeleteNamespace(namespace string) error {
	return k8s.clientSet.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
}

// GetResourcesInRelease returns all runtime resources under a given release name in a namespace
func (k8s *K8sClient) GetResourcesInRelease(releaseName string, namespace string) *ReleaseResources {
	rr := NewReleaseResources(releaseName)
	listOpts := metav1.ListOptions{}

	// Deployments
	deployList, err := k8s.clientSet.AppsV1().Deployments(namespace).List(context.TODO(), listOpts)
	if err != nil {
		panic("Error getting deployments in namespace " + namespace)
	}
	for _, deployment := range deployList.Items {
		val, exists := deployment.Labels[LabelReleaseName]
		if exists && val == releaseName {
			rr.Deployments = append(rr.Deployments, deployment)
		}
	}
	// StatefulSets
	ssList, err := k8s.clientSet.AppsV1().StatefulSets(namespace).List(context.TODO(), listOpts)
	if err != nil {
		panic("Error getting statefulsets in namespace " + namespace)
	}
	for _, ss := range ssList.Items {
		val, exists := ss.Labels[LabelReleaseName]
		if exists && val == releaseName {
			rr.StatefulSets = append(rr.StatefulSets, ss)
		}
	}
	// Daemonsets
	dsList, err := k8s.clientSet.AppsV1().DaemonSets(namespace).List(context.TODO(), listOpts)
	if err != nil {
		panic("Error getting daemonsets in namespace " + namespace)
	}
	for _, ds := range dsList.Items {
		val, exists := ds.Labels[LabelReleaseName]
		if exists && val == releaseName {
			rr.DaemonSets = append(rr.DaemonSets, ds)
		}
	}
	// Jobs
	jobsList, err := k8s.clientSet.BatchV1().Jobs(namespace).List(context.TODO(), listOpts)
	if err != nil {
		panic("Error getting jobs in namespace " + namespace)
	}
	for _, job := range jobsList.Items {
		val, exists := job.Labels[LabelReleaseName]
		if exists && val == releaseName {
			rr.Jobs = append(rr.Jobs, job)
		}
	}
	return rr
}

// WaitForRelease pauses for up to 'timeout' seconds waiting for the specified release to be fully installed
func (k8s *K8sClient) WaitForRelease(releaseName string, namespace string, timeout time.Duration) bool {
	start := time.Now()
	rr := k8s.GetResourcesInRelease(releaseName, namespace)
	for !rr.IsReleaseInstalled() {
		time.Sleep(2 * time.Second)
		end := time.Now()
		elapsed := end.Sub(start)
		if elapsed > timeout {
			return false
		}
		rr = k8s.GetResourcesInRelease(releaseName, namespace)
		logrus.Debugf("Waiting for release %v [%v] Elapsed=%v\n", releaseName, namespace, elapsed)
	}
	return true
}
