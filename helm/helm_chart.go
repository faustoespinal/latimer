package helm

import (
	"encoding/json"
	"fmt"
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

	// The chart descriptor
	Descriptor *core.ChartDescriptor `json:"descriptor"`

	// The loaded values map
	ValuesMap map[string]interface{}
}

// NewChart creates a new instance of a helm chart
func NewChart(chartDescriptor *core.ChartDescriptor) *Chart {
	hc := new(Chart)
	hc.Name = chartDescriptor.Name
	hc.ChartRef = chartDescriptor.ChartLocator
	hc.Descriptor = chartDescriptor
	valuesFiles := make([]string, 0)
	for _, valueFile := range hc.Descriptor.Values {
		valuesFiles = append(valuesFiles, valueFile.URL)
	}
	valMap, err := loadHelmValues(valuesFiles)
	if err != nil {
		panic("Error loading values file for chart: " + hc.Name)
	}
	hc.ValuesMap = valMap
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
	releaseInfo, err := helmClient.Install(releaseName, releaseNamespace, hc.ChartRef, hc.ValuesMap)
	status := true
	if releaseInfo != nil && err != nil {
		logrus.Warningf("Helm chart %v is already installed in the namespace %v", releaseName, releaseNamespace)
		status = false
	} else if err != nil {
		logrus.Errorf("Install failed [%v]", err)
		status = false
	} else {
		fmt.Printf("%v", releaseInfo.Info.Notes)
		fmt.Printf("Helm chart %v installed to namespace %v\n", releaseName, releaseNamespace)
		fmt.Println("----------------------------------------------------------------------------------------")
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
			logrus.Errorf("Delete failed [%v]", err.Error())
			status = false
		} else {
			fmt.Printf("Helm chart %v deleted from namespace %v\n", releaseName, releaseNamespace)
		}
	}
	return status
}

// Status returns the status of the  installation
func (hc *Chart) Status(sc *core.SystemContext) kube.InstallStatus {
	k8s := sc.Context.KubeClient
	namespace := hc.Descriptor.Namespace
	releaseName := hc.Descriptor.ReleaseName
	rr, err := k8s.GetResourcesInRelease(releaseName, namespace)
	if err != nil {
		return kube.InstallationError
	}
	return rr.ReleaseStatus()
}
