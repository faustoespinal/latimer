package helm

import (
	"testing"
)

const (
	ChartRef = "stable/memcached"
)

func Test_helm_pull(t *testing.T) {
	t.Run("helm-pull-remote", func(t *testing.T) {
		t.Logf("Testing helm pull from remote registry")

		helmClient := NewHelmClient()
		outPath, err := helmClient.Pull(ChartRef, "/tmp")
		if err != nil {
			panic(err.Error())
		}
		t.Logf("Pulled helm chart to path: %v\n", outPath)
	})
}

func Test_helm_install(t *testing.T) {
	t.Run("helm-install-remote", func(t *testing.T) {
		t.Logf("Testing helm install from remote registry")

		helmClient := NewHelmClient()
		releaseName := "test-mem"
		namespace := "paas"
		err := helmClient.Install(releaseName, namespace, ChartRef)
		if err != nil {
			panic(err.Error())
		}
		t.Logf("Installed helm chart [%v] to namespace: %v\n", ChartRef, namespace)
	})
}

func Test_helm_delete(t *testing.T) {
	t.Run("helm-delete", func(t *testing.T) {
		t.Logf("Testing helm delete functionality")

		releaseName := "test-mem"
		namespace := "paas"

		helmClient := NewHelmClient()
		err := helmClient.Delete(releaseName, namespace)
		if err != nil {
			panic(err.Error())
		}
		t.Logf("Deleted helm chart [%v] from namespace: %v\n", releaseName, namespace)
	})
}
