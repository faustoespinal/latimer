package manifest

import (
	"latimer/core"
	"os/user"
	"path/filepath"
	"testing"
)

const (
	ManifestFilePath = "../test/install-manifest-3.yaml"
)

// Returns an initialized system context
func getSystemContext(name string) *core.SystemContext {
	lc := core.GetLatimerContext()
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	kubeConfigPath := filepath.Join(user.HomeDir, ".kube", "config")
	lc.InitLatimer(kubeConfigPath, ManifestFilePath, []string{})
	sc := new(core.SystemContext)
	sc.Context = lc
	sc.Name = name
	sc.WorkTempDir = lc.LatimerTempDir
	return sc
}

func Test_ManifestCapabilities(t *testing.T) {
	t.Run("manifest-install-order", func(t *testing.T) {
		values := map[string]string{}
		m, err := NewManifest(ManifestFilePath, values)
		if err != nil {
			t.Errorf("%v", err)
		}
		installationTable := map[string]bool{}
		manifestID := m.GetID()
		installList := m.createInstallOrderFrom(core.InstallableItem{
			Name: manifestID,
			Kind: core.ManifestType,
		}, installationTable)
		t.Logf("Install Order List: %v", installList)
	})

	t.Run("manifest-installation", func(t *testing.T) {
		values := map[string]string{}
		m, err := NewManifest(ManifestFilePath, values)
		if err != nil {
			t.Errorf("%v", err)
		}

		sc := getSystemContext(m.GetID())
		status := m.Install(sc)
		if !status {
			t.Errorf("Installation failed %v", m.Descriptor)
		}
		t.Logf("==================== Manifest installed: %v ======================", status)
	})

	t.Run("manifest-deletion", func(t *testing.T) {
		values := map[string]string{}
		m, err := NewManifest(ManifestFilePath, values)
		if err != nil {
			t.Errorf("%v", err)
		}

		sc := getSystemContext(m.GetID())
		status := m.Uninstall(sc)
		if !status {
			t.Errorf("Uninstallation failed %v", m.Descriptor)
		}
		t.Logf("==================== Manifest uninstalled: %v ======================", status)
	})
}
