package core

import (
	"bufio"
	"bytes"
	"html/template"
	"log"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	// ChartType is constant denoting an installable item of type chart
	ChartType = "chart"
	// PackageType is constant denoting an installable item of type package
	PackageType = "package"
	// ManifestType is constant denoting an installable item of type manifest
	ManifestType = "manifest"
)

// ChartDescriptor describes a chart
type ChartDescriptor struct {
	Name         string `json:"name"`
	ChartName    string `json:"chartName" yaml:"chartName"`
	Namespace    string `json:"namespace"`
	ChartLocator string `json:"chartLocator" yaml:"chartLocator"`
	ReleaseName  string `json:"releaseName" yaml:"releaseName"`
	// Timeout is the value in seconds to wait for chart to come up before giving up
	Timeout int `json:"timeout,omitempty"`
	Values  []struct {
		// URL is the locator for the values yaml file
		URL string `json:"url"`
	} `json:"values,omitempty"`
}

// PackageDescriptor groups a collection of chart descriptors
type PackageDescriptor struct {
	Name   string            `json:"name"`
	Charts []InstallableItem `json:"charts"`
}

// ManifestDescriptor describes collection of packages and charts to be installed
type ManifestDescriptor struct {
	Metadata        InstallableItem     `json:"metadata"`
	Charts          []ChartDescriptor   `json:"charts"`
	Packages        []PackageDescriptor `json:"packages,omitempty"`
	DependencyItems []struct {
		Name     string            `json:"name"`
		Requires []InstallableItem `json:"requires"`
	} `json:"dependencies" yaml:"dependencies"`
}

// LoadManifestDescriptor creates a new manifest descriptor object from file contents
func LoadManifestDescriptor(filePath string, values map[string]string) (*ManifestDescriptor, error) {
	m := new(ManifestDescriptor)

	logrus.Infof("Templating manifest file with args: [%v]", values)
	tpl, err := template.ParseFiles(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	var b bytes.Buffer // A Buffer needs no initialization.
	wr := bufio.NewWriter(&b)
	tpl.Execute(wr, values)
	wr.Flush()

	yamlBytes := b.Bytes()
	err = yaml.Unmarshal(yamlBytes, m)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return nil, err
	}
	dirname := filepath.Dir(filePath)
	for _, chart := range m.Charts {
		if len(chart.Values) > 0 {
			for idx, path := range chart.Values {
				chart.Values[idx].URL = filepath.Join(dirname, path.URL)
			}
		}
	}
	return m, nil
}
