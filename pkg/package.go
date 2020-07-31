package pkg

import (
	"encoding/json"
	"latimer/core"
	"latimer/helm"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Package class describing collection of packages and charts to be installed
type Package struct {
	Name         string             `json:"name"`
	Installables []core.Installable `json:"charts"`
}

// NewPackage creates a package object
func NewPackage(pd *core.PackageDescriptor, chartRef string) *Package {
	p := new(Package)
	p.Name = pd.Name
	p.Installables = make([]core.Installable, 0)
	for _, item := range pd.Charts {
		switch item.Kind {
		case core.ChartType:
			hc := helm.NewChart(item.Name, chartRef)
			p.Installables = append(p.Installables, hc)
		case core.PackageType:
		case core.ManifestType:
		}
	}
	return p
}

// Return string representation of package
func (p *Package) String() string {
	packageBytes, err := json.Marshal(p)
	if err != nil {
		logrus.Error("Generating JSON representation")
	}
	return string(packageBytes)
}

// StringYaml returns yaml representation of a package
func (p *Package) StringYaml() string {
	packageBytes, err := yaml.Marshal(p)
	if err != nil {
		logrus.Error("Generating YAML representation")
	}
	return string(packageBytes)
}

// Install the contents of the installable
func (p *Package) Install(sc *core.SystemContext) bool {
	for _, swItem := range p.Installables {
		swItem.Install(sc)
	}
	return true
}

// Uninstall the contents of this installable
func (p *Package) Uninstall(sc *core.SystemContext) bool {
	for _, swItem := range p.Installables {
		swItem.Uninstall(sc)
	}
	return true
}

// Status returns the status of the  installation
func (p *Package) Status() core.InstallStatus {
	return core.Ready
}

// GetID returns the identifier name for this Installable.
func (p *Package) GetID() string {
	return p.Name
}
