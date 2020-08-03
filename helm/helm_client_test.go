package helm

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	ChartRef = "stable/memcached"
)

func Test_helm_pull(t *testing.T) {
	t.Run("helm-pull-remote", func(t *testing.T) {
		t.Logf("Testing helm pull from remote registry")

		tmpDir, err := ioutil.TempDir(os.TempDir(), "helm-*")
		if err != nil {
			t.Errorf("Error creating temp directory %v %v", err.Error(), tmpDir)
			panic(err.Error())
		}
		defer os.RemoveAll(tmpDir)

		helmClient := NewHelmClient()
		outPath, err := helmClient.Pull(ChartRef, tmpDir)
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
		valuesMap := map[string]interface{}{}
		release, err := helmClient.Install(releaseName, namespace, ChartRef, valuesMap)
		if err != nil {
			panic(err.Error())
		}
		t.Logf("Installed helm chart [%v] to namespace: %v\n", release.Name, release.Namespace)
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
