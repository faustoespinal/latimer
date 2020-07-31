package kube

import (
	"os/user"
	"path/filepath"
	"testing"
	"time"
)

const (
	NAMESPACE = "test-ns"
)

func getDefaultKubeConfigPath() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	defaultKubeConfigPath := filepath.Join(user.HomeDir, ".kube", "config")
	return defaultKubeConfigPath
}

func Test_CreateDeleteNamespace(t *testing.T) {
	k8s, err := NewK8sClient(getDefaultKubeConfigPath())
	if err != nil {
		panic(err)
	}

	t.Run("create-namespace", func(t *testing.T) {
		ns, err := k8s.CreateNamespace(NAMESPACE)
		if err != nil {
			t.Errorf("Error creating namespace called %v\n", NAMESPACE)
		} else {
			t.Logf("Namespace %v created: %v\n", NAMESPACE, *ns)
		}
	})
	time.Sleep(1000 * time.Millisecond)
	t.Run("namespace-exists", func(t *testing.T) {
		ns := k8s.GetNamespace(NAMESPACE)
		if ns != nil {
			t.Logf("Namespace %v found: %v\n", NAMESPACE, *ns)
		} else {
			t.Errorf("Expecting a namespace called %v\n", NAMESPACE)
		}
	})
	time.Sleep(1000 * time.Millisecond)
	t.Run("delete-namespace", func(t *testing.T) {
		err := k8s.DeleteNamespace(NAMESPACE)
		if err != nil {
			t.Errorf("Error deleting namespace called %v\n", NAMESPACE)
		} else {
			t.Logf("Namespace %v deleted\n", NAMESPACE)
		}
	})
}

func Test_GetResourcesInRelease(t *testing.T) {
	k8s, err := NewK8sClient(getDefaultKubeConfigPath())
	if err != nil {
		panic(err)
	}

	releaseName := "test-mysql"
	namespace := "paas"
	t.Run("deployments-in", func(t *testing.T) {
		resources := k8s.GetResourcesInRelease(releaseName, namespace)
		if len(resources.Deployments) <= 0 {
			t.Errorf("No deployments in namespace %v\n", namespace)
		} else {
			for _, d := range resources.Deployments {
				t.Logf("Name: %v -- AvailableReplicas: %v, ReadyReplicas: %v\n", d.Name, d.Status.AvailableReplicas, d.Status.ReadyReplicas)
			}
			t.Logf("Found %v deployments in namespace %v\n", len(resources.Deployments), namespace)
		}
	})
}

func Test_WaitForRelease(t *testing.T) {
	k8s, err := NewK8sClient(getDefaultKubeConfigPath())
	if err != nil {
		panic(err)
	}

	releaseName := "test-mysql"
	namespace := "paas"
	t.Run("wait-for-release", func(t *testing.T) {
		waitStatus := k8s.WaitForRelease(releaseName, namespace, 3*time.Second)
		if waitStatus {
			t.Logf("Release %v has been installed\n", releaseName)
		} else {
			t.Errorf("Timeout expired for release %v in namespace %v\n", releaseName, namespace)
		}
	})
}
