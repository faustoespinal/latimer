package manifest

import (
	"encoding/json"
	"fmt"
	"latimer/core"
	"latimer/helm"
	"log"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	// MaxInstallableWaitTime is the maximum amount of time to wait for an Installable item
	MaxInstallableWaitTime = (8 * time.Minute)
)

// Manifest describes collection of packages and charts to be installed
type Manifest struct {
	// The descriptor of the manifest
	Descriptor *core.ManifestDescriptor

	charts       map[string]*core.ChartDescriptor
	packages     map[string]*core.PackageDescriptor
	dependencies map[string][]core.InstallableItem
}

// NewManifest creates a new manifest object from file contents
func NewManifest(manifestArgs core.ManifestInput) (*Manifest, error) {
	m := new(Manifest)

	descriptor, err := core.LoadManifestDescriptor(manifestArgs)
	if err != nil {
		return nil, err
	}
	m.Descriptor = descriptor
	m.charts = map[string]*core.ChartDescriptor{}
	m.packages = map[string]*core.PackageDescriptor{}
	m.dependencies = map[string][]core.InstallableItem{}

	manifestID := m.GetID()
	// Index the charts by name into a map
	manifestDeps := make([]core.InstallableItem, 0)
	for idx, c := range descriptor.Charts {
		m.charts[c.Name] = &(descriptor.Charts[idx])
		manifestDeps = append(manifestDeps, core.InstallableItem{
			Name: c.Name,
			Kind: core.ChartType,
		})
	}
	// Index the packages by name into a map
	for _, p := range descriptor.Packages {
		m.packages[p.Name] = &p
		manifestDeps = append(manifestDeps, core.InstallableItem{
			Name: p.Name,
			Kind: core.PackageType,
		})
	}
	// Index the dependencies
	for _, dItem := range descriptor.DependencyItems {
		m.dependencies[dItem.Name] = dItem.Requires
	}
	m.dependencies[manifestID] = manifestDeps
	return m, nil
}

// GetID returns the identifier name for this Installable.
func (m *Manifest) GetID() string {
	return m.Descriptor.Metadata.Name
}

// Return string representation of manifest
func (m *Manifest) String() string {
	manifestBytes, err := json.Marshal(m)
	if err != nil {
		logrus.Error("Generating JSON representation")
	}
	return string(manifestBytes)
}

// StringYaml returns yaml representation of a manifest
func (m *Manifest) StringYaml() string {
	manifestBytes, err := yaml.Marshal(m)
	if err != nil {
		logrus.Error("Generating YAML representation")
	}
	return string(manifestBytes)
}

// Install the contents of the installable
func (m *Manifest) Install(sc *core.SystemContext) bool {
	manifestID := m.GetID()
	// Initialize installation buffer table
	installationTable := map[string]bool{}
	installList := m.createInstallOrderFrom(core.InstallableItem{
		Name: manifestID,
		Kind: core.ManifestType,
	}, installationTable)
	fmt.Printf("Installing manifest: %v\n", m.Descriptor.Metadata.Name)
	for _, installItem := range installList {
		logrus.Infof("%% name=%v type=%v\n", installItem.Name, installItem.Kind)
		sysCtxt := *sc
		switch installItem.Kind {
		case core.ChartType:
			c := m.charts[installItem.Name]
			hc := helm.NewChart(c.Name, c.ChartLocator)
			// Clone the system context and override values.
			sysCtxt.DeploymentSpace = c.Namespace
			sysCtxt.ReleaseName = c.ReleaseName
			fmt.Printf("    Installing chart: %v\n", hc.Name)
			hc.Install(&sysCtxt)
			k8sClient := sysCtxt.Context.KubeClient
			k8sClient.WaitForRelease(sysCtxt.ReleaseName, sysCtxt.DeploymentSpace, MaxInstallableWaitTime)
			logrus.Infof("Installed HELM chart %v\n", sysCtxt.ReleaseName)
		case core.PackageType:
			p := m.packages[installItem.Name]
			fmt.Printf("    Installing package: %v\n", p.Name)
			//pkg := pkg.NewPackage(p, chartRef)
			logrus.Infof("@@@@@ Package %v\n", p.Name)
		case core.ManifestType:
			logrus.Infof("Installed manifest %v\n", installItem.Name)
		}
		logrus.Info("===============================================================")
	}
	return true
}

// Uninstall the contents of this installable
func (m *Manifest) Uninstall(sc *core.SystemContext) bool {
	manifestID := m.GetID()
	// Initialize installation buffer table
	installationTable := map[string]bool{}
	installList := m.createInstallOrderFrom(core.InstallableItem{
		Name: manifestID,
		Kind: core.ManifestType,
	}, installationTable)
	for idx := len(installList) - 1; idx >= 0; idx-- {
		installItem := installList[idx]
		logrus.Infof("%% name=%v type=%v\n", installItem.Name, installItem.Kind)
		switch installItem.Kind {
		case core.ChartType:
			c := m.charts[installItem.Name]
			urlRef, err := url.Parse(c.ChartLocator)
			if err != nil {
				panic(err)
			}
			filePath := urlRef.RequestURI()
			hc := helm.NewChart(c.Name, filePath)
			// Clone the system context and override values.
			sysCtxt := *sc
			sysCtxt.DeploymentSpace = c.Namespace
			sysCtxt.ReleaseName = c.ReleaseName
			hc.Uninstall(&sysCtxt)
			logrus.Infof("Uninstalled HELM chart %v\n", sysCtxt.ReleaseName)
		case core.PackageType:
			p := m.packages[installItem.Name]
			logrus.Infof("@@@@@ Package %v\n", p.Name)
		case core.ManifestType:
			logrus.Infof("Installed manifest %v\n", installItem.Name)
		}
		log.Println("===============================================================")
	}
	return true
}

// Status returns the status of the  installation
func (m *Manifest) Status() core.InstallStatus {
	return core.Ready
}

// Creates an ordered list of installation items reflecting the installation order given dependencies
func (m *Manifest) createInstallOrderFrom(installItem core.InstallableItem, installTable map[string]bool) []core.InstallableItem {
	name := installItem.Name
	installed := installTable[name]

	if installed {
		return nil
	}
	deps, found := m.dependencies[name]
	installList := make([]core.InstallableItem, 0)
	if found {
		for _, d := range deps {
			depInstall := m.createInstallOrderFrom(d, installTable)
			if depInstall != nil {
				installList = append(installList, depInstall...)
			}
		}
	}
	installList = append(installList, installItem)
	installTable[name] = true
	return installList
}
