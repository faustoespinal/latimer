package kube

import (
	"os/user"
	"path/filepath"
	"testing"
	"time"
)

func getDefaultKubeConfigPath() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	defaultKubeConfigPath := filepath.Join(user.HomeDir, ".kube", "config")
	return defaultKubeConfigPath
}

func Test_GetNamespace(t *testing.T) {
	k8s, err := NewK8sClient(getDefaultKubeConfigPath())
	if err != nil {
		panic(err)
	}

	namespace := "paas"
	t.Run("success", func(t *testing.T) {
		ns := k8s.GetNamespace(namespace)
		if ns != nil {
			t.Logf("Namespace %v found: %v\n", namespace, *ns)
		} else {
			t.Errorf("Expecting a namespace called %v\n", namespace)
		}
	})
}

func Test_CreateDeleteNamespace(t *testing.T) {
	k8s, err := NewK8sClient(getDefaultKubeConfigPath())
	if err != nil {
		panic(err)
	}

	namespace := "test-ns"
	t.Run("create-ns", func(t *testing.T) {
		ns, err := k8s.CreateNamespace(namespace)
		if err != nil {
			t.Errorf("Error creating namespace called %v\n", namespace)
		} else {
			t.Logf("Namespace %v created: %v\n", namespace, *ns)
		}
	})
	t.Run("delete-ns", func(t *testing.T) {
		err := k8s.DeleteNamespace(namespace)
		if err != nil {
			t.Errorf("Error deleting namespace called %v\n", namespace)
		} else {
			t.Logf("Namespace %v deleted\n", namespace)
		}
	})
	// time.Sleep(2000 * time.Millisecond)
	// t.Run("check-no-ns", func(t *testing.T) {
	// 	ns := k8s.GetNamespace(namespace)
	// 	if ns == nil {
	// 		t.Logf("Namespace %v not found\n", namespace)
	// 	} else {
	// 		t.Errorf("Namespace %v still exists and should not be: %v\n", namespace, *ns)
	// 	}
	// })
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
