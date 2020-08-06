package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"latimer/core"
	"latimer/helm"
	"latimer/kube"
	"latimer/pkg"
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

	charts       map[string]*helm.Chart
	packages     map[string]*pkg.Package
	dependencies map[string][]core.InstallableItem
}

// NewManifest creates a new manifest object from file contents
func NewManifest(filePath string, values map[string]string) (*Manifest, error) {
	m := new(Manifest)

	descriptor, err := core.LoadManifestDescriptor(filePath, values)
	if err != nil {
		return nil, err
	}
	m.Descriptor = descriptor
	m.charts = map[string]*helm.Chart{}
	m.packages = map[string]*pkg.Package{}
	m.dependencies = map[string][]core.InstallableItem{}

	manifestID := m.GetID()
	manifestDeps := make([]core.InstallableItem, 0)
	// Index the charts by name into a map
	for idx, c := range descriptor.Charts {
		m.charts[c.Name] = helm.NewChart(&(descriptor.Charts[idx]))
		manifestDeps = append(manifestDeps, core.InstallableItem{
			Name: c.Name,
			Kind: core.ChartType,
		})
	}
	// Index the packages by name into a map
	for _, p := range descriptor.Packages {
		charts := make([]*helm.Chart, 0)
		for _, pkgChart := range p.Charts {
			name := pkgChart.Name
			helmChart, found := m.charts[name]
			if found {
				charts = append(charts, helmChart)
			}
		}
		m.packages[p.Name] = pkg.NewPackage(&p, charts)
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
	installList := m.installList()
	fmt.Printf("Installing manifest: %v [%v]\n", m.Descriptor.Metadata.Name, installList)
	for _, installItem := range installList {
		// Clone the system context and override values.
		sysCtxt := *sc
		m.waitForDependencies(&sysCtxt, installItem.Name)
		switch installItem.Kind {
		case core.ChartType:
			hc, found := m.charts[installItem.Name]
			if !found {
				panic("Unrecognized chart: " + installItem.Name)
			}
			c := hc.Descriptor
			releaseName := c.ReleaseName
			fmt.Printf("    Installing chart: %v", hc.Name)
			hc.Install(&sysCtxt)
			logrus.Infof("Installed HELM chart %v", releaseName)
		case core.PackageType:
			p := m.packages[installItem.Name]
			fmt.Printf("    Installing package: %v", p.Name)
			p.Install(&sysCtxt)
			logrus.Infof("Installed Package %v", p.Name)
		case core.ManifestType:
			logrus.Infof("Installed manifest %v", installItem.Name)
		}
		logrus.Info("===============================================================")
	}
	return true
}

// Uninstall the contents of this installable
func (m *Manifest) Uninstall(sc *core.SystemContext) bool {
	manifestID := m.GetID()
	installList := m.installList()
	logrus.Infof("Uninstall manifest %v : [%v]", manifestID, installList)
	for idx := len(installList) - 1; idx >= 0; idx-- {
		installItem := installList[idx]
		sysCtxt := *sc
		logrus.Infof("Uninstalling item: %v %v", installItem.Name, installItem.Kind)
		switch installItem.Kind {
		case core.ChartType:
			hc := m.charts[installItem.Name]
			c := hc.Descriptor
			releaseName := c.ReleaseName
			hc.Uninstall(&sysCtxt)
			logrus.Infof("Uninstalled HELM chart %v", releaseName)
		case core.PackageType:
			p := m.packages[installItem.Name]
			p.Uninstall(&sysCtxt)
			logrus.Infof("Uninstalled Package %v", p.Name)
		case core.ManifestType:
			logrus.Infof("Uninstalled manifest %v", installItem.Name)
		}
		logrus.Infof("===============================================================")
	}
	return true
}

// Status returns the status of the  installation
func (m *Manifest) Status(sc *core.SystemContext) kube.InstallStatus {
	for _, swItem := range m.charts {
		chartSC := *sc
		status := swItem.Status(&chartSC)
		if status != kube.Ready {
			return kube.NotReady
		}
	}
	return kube.Ready
}

// Wait for all dependencies before installing the given itemID
func (m *Manifest) waitForDependencies(sc *core.SystemContext, itemID string) error {
	depItems, found := m.dependencies[itemID]
	if found {
		// Default 5 minutes
		timeout := 300 * time.Second
		for _, item := range depItems {
			var installable core.Installable = nil
			chart, foundChart := m.charts[item.Name]
			if foundChart {
				installable = chart
			} else {
				pkg, foundPkg := m.charts[item.Name]
				if foundPkg {
					installable = pkg
				}
			}
			if installable != nil {
				fmt.Printf("    %v waiting for dependency %v to complete install", itemID, installable.GetID())
				start := time.Now()
				for installable.Status(sc) != kube.Ready {
					time.Sleep(2 * time.Second)
					end := time.Now()
					elapsed := end.Sub(start)
					if elapsed > timeout {
						return errors.New("Timeout expired for: " + installable.GetID())
					}
					logrus.Debugf("       Waiting for release %v Elapsed=%v", installable.GetID(), elapsed)
				}
			}
		}
	}
	return nil
}

// followDeps creates an ordered list of installation items reflecting dependencies from the given root
func (m *Manifest) followDeps(installItem core.InstallableItem, installTable map[string]bool) []core.InstallableItem {
	name := installItem.Name
	installed := installTable[name]

	if installed {
		return nil
	}
	deps, found := m.dependencies[name]
	installList := make([]core.InstallableItem, 0)
	if found {
		for _, d := range deps {
			depInstall := m.followDeps(d, installTable)
			if depInstall != nil {
				installList = append(installList, depInstall...)
			}
		}
	}
	installList = append(installList, installItem)
	installTable[name] = true
	if installItem.Kind == core.PackageType {
		// Mark all contained
		pkgName := installItem.Name
		pkg, found := m.packages[pkgName]
		if found {
			for _, c := range pkg.Charts {
				installTable[c.Name] = true
			}
		}
	}
	return installList
}

// Creates an ordered list of installation items reflecting the installation order given dependencies
func (m *Manifest) installList() []core.InstallableItem {
	installTable := make(map[string]bool, 0)
	installList := make([]core.InstallableItem, 0)

	for k := range m.dependencies {
		list := []core.InstallableItem(nil)
		if _, found := m.charts[k]; found {
			list = m.followDeps(core.InstallableItem{
				Name: k,
				Kind: core.ChartType,
			}, installTable)
		} else if _, found := m.packages[k]; found {
			list = m.followDeps(core.InstallableItem{
				Name: k,
				Kind: core.PackageType,
			}, installTable)
		}
		installList = append(installList, list...)
	}
	list := m.followDeps(core.InstallableItem{
		Name: m.GetID(),
		Kind: core.ManifestType,
	}, installTable)
	installList = append(installList, list...)
	return installList
}
