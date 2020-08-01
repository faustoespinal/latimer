package helm

import (
	"latimer/core"
	"testing"
)

const (
	MemChartRef = "stable/memcached"
)

var chartDescriptor core.ChartDescriptor = core.ChartDescriptor{
	Name:         "sample-chart",
	ChartName:    "sample-chart",
	ChartLocator: "stable/memcached",
	Namespace:    "paas",
	ReleaseName:  "test-memcached",
	Values: []struct {
		URL string `json:"url"`
	}{},
}

func Test_helmchart(t *testing.T) {
	t.Run("helm-chart-create", func(t *testing.T) {
		t.Logf("Testing helm chart create")

		chart := NewChart(&chartDescriptor)
		if chart.Name != chartDescriptor.Name {
			t.Errorf("Chart creation failed %v  descriptor=%v", chart, chartDescriptor.Name)
		}
		t.Logf("Created helm chart: %v\n", chart)
	})
}
