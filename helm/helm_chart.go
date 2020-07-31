package helm

import (
	"encoding/json"
	"latimer/core"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Chart class is a wrapper around a k8s HELM chart
type Chart struct {
	// Name of the chart
	Name string `json:"name"`

	// ChartRef is the locator for the chart (eg chart repository or local file system)
	ChartRef string `json:"chartRef"`
}

// NewChart creates a new instance of a helm chart
func NewChart(name string, chartRef string) *Chart {
	hc := new(Chart)
	hc.Name = name
	hc.ChartRef = chartRef
	return hc
}

// GetID returns the identifier name for this Installable.
func (hc *Chart) GetID() string {
	return hc.Name
}

// Return string representation of chart
func (hc *Chart) String() string {
	chartBytes, err := json.Marshal(hc)
	if err != nil {
		logrus.Error("Error generating JSON representation")
	}
	return string(chartBytes)
}

// StringYaml returns yaml representation of a chart
func (hc *Chart) StringYaml() string {
	chartBytes, err := yaml.Marshal(hc)
	if err != nil {
		logrus.Error("Error generating YAML representation")
	}
	return string(chartBytes)
}

// Install the contents of the installable
func (hc *Chart) Install(sc *core.SystemContext) bool {
	releaseNamespace := sc.DeploymentSpace
	releaseName := sc.ReleaseName

	helmClient := NewHelmClient()
	releaseInfo, err := helmClient.Install(releaseName, releaseNamespace, hc.ChartRef)
	status := true
	if releaseInfo != nil && err != nil {
		logrus.Errorf("Helm chart %v is already installed in the namespace %v", releaseName, releaseNamespace)
		status = false
	} else if err != nil {
		logrus.Errorf("Install failed [%v]\n", err.Error())
		status = false
	}
	return status
}

// Uninstall the contents of this installable
func (hc *Chart) Uninstall(sc *core.SystemContext) bool {
	releaseNamespace := sc.DeploymentSpace
	releaseName := sc.ReleaseName

	helmClient := NewHelmClient()
	err := helmClient.Delete(releaseName, releaseNamespace)
	status := true
	if err != nil {
		logrus.Errorf("Delete failed [%v]\n", err.Error())
		status = false
	}
	return status
}

// Status returns the status of the  installation
func (hc *Chart) Status() core.InstallStatus {
	return core.Ready
}
