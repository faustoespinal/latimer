package helm

import (
	"encoding/json"
	"latimer/core"
	"latimer/kube"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Chart class is a wrapper around a k8s HELM chart
type Chart struct {
	// Name of the chart
	Name string `json:"name"`

	// ChartRef is the locator for the chart (eg chart repository or local file system)
	ChartRef string `json:"chartRef"`

	Descriptor *core.ChartDescriptor `json:"descriptor"`
}

// NewChart creates a new instance of a helm chart
func NewChart(chartDescriptor *core.ChartDescriptor) *Chart {
	hc := new(Chart)
	hc.Name = chartDescriptor.Name
	hc.ChartRef = chartDescriptor.ChartLocator
	hc.Descriptor = chartDescriptor
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
	releaseNamespace := hc.Descriptor.Namespace
	releaseName := hc.Descriptor.ReleaseName

	helmClient := NewHelmClient()
	releaseInfo, err := helmClient.Install(releaseName, releaseNamespace, hc.ChartRef)
	status := true
	if releaseInfo != nil && err != nil {
		logrus.Warningf("Helm chart %v is already installed in the namespace %v", releaseName, releaseNamespace)
		status = false
	} else if err != nil {
		logrus.Errorf("Install failed [%v]\n", err.Error())
		status = false
	}
	return status
}

// Uninstall the contents of this installable
func (hc *Chart) Uninstall(sc *core.SystemContext) bool {
	releaseNamespace := hc.Descriptor.Namespace
	releaseName := hc.Descriptor.ReleaseName

	status := true
	helmClient := NewHelmClient()
	release, err := helmClient.Status(releaseName, releaseNamespace)

	// If release does not exist already we just return successful uninstall
	if err == nil && release != nil {
		err := helmClient.Delete(releaseName, releaseNamespace)
		if err != nil {
			logrus.Errorf("Delete failed [%v]\n", err.Error())
			status = false
		}
	}
	return status
}

// Status returns the status of the  installation
func (hc *Chart) Status(sc *core.SystemContext) kube.InstallStatus {
	k8s := sc.Context.KubeClient
	namespace := hc.Descriptor.Namespace
	releaseName := hc.Descriptor.ReleaseName
	rr := k8s.GetResourcesInRelease(releaseName, namespace)
	return rr.ReleaseStatus()
}
