package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

// HelmClient represents a helm client capable of issuing helm commands againts a kubernetes API server in a given
// namespace
type HelmClient struct {
}

// NewHelmClient creates a new helm client to manage charts in a specified namespace
func NewHelmClient() *HelmClient {
	hc := new(HelmClient)
	return hc
}

// Status returns the status of the helm release in the given namespace.  Used to know if release exists
func (hc *HelmClient) Status(releaseName string, namespace string) (*release.Release, error) {
	actionConfig, err := newHelmConfig(namespace)
	if err != nil {
		return nil, err
	}

	iCli := action.NewStatus(actionConfig)
	releaseInfo, err := iCli.Run(releaseName)
	return releaseInfo, err
}

// Upgrade performs a 'helm upgrade' on the specified release name
func (hc *HelmClient) Upgrade(releaseName string, namespace string, chartRef string, valuesMap map[string]interface{}) (*release.Release, error) {
	chart, err := hc.loadChart(chartRef)
	if err != nil {
		logrus.Errorf("Error loading chart from location=%v", chartRef)
		return nil, err
	}

	actionConfig, err := newHelmConfig(namespace)
	if err != nil {
		return nil, err
	}

	iCli := action.NewUpgrade(actionConfig)
	releaseInfo, err := iCli.Run(releaseName, chart, valuesMap)
	return releaseInfo, err
}

// Install deploys the helm chart located in the specified chart location
func (hc *HelmClient) Install(releaseName string, namespace string, chartRef string, valuesMap map[string]interface{}) (*release.Release, error) {
	// Check if release name is already present
	releaseInfo, err := hc.Status(releaseName, namespace)
	if releaseInfo != nil && err == nil {
		//return releaseInfo, fmt.Errorf("Release %v exists in namespace %v", releaseName, namespace)
		logrus.Warningf("Release name %v exists in namespace %v, will upgrade", releaseName, namespace)
		return hc.Upgrade(releaseName, namespace, chartRef, valuesMap)
	}
	logrus.Infof("Installing chart with chartRef=%v to namespace %v", chartRef, namespace)
	chart, err := hc.loadChart(chartRef)
	if err != nil {
		logrus.Errorf("Error loading chart from location=%v", chartRef)
		return nil, err
	}

	actionConfig, err := newHelmConfig(namespace)
	if err != nil {
		return nil, err
	}

	iCli := action.NewInstall(actionConfig)
	iCli.Namespace = namespace
	iCli.ReleaseName = releaseName
	iCli.DryRun = false

	rel, err := iCli.Run(chart, valuesMap)
	if err != nil {
		logrus.Errorf("Error running install: [%v]", err)
		return nil, err
	}
	logrus.Debugf("Successfully submitted release: %v --> %v\n", rel.Name, rel.Namespace)
	return rel, nil
}

// Delete installs the helm chart located in the specified chart path location
func (hc *HelmClient) Delete(releaseName string, namespace string) error {
	actionConfig, err := newHelmConfig(namespace)
	if err != nil {
		return err
	}

	iCli := action.NewUninstall(actionConfig)

	uninstallResp, err := iCli.Run(releaseName)
	if err != nil {
		return err
	}
	logrus.Debugf("Uninstalled %v: [%v]\n", releaseName, uninstallResp.Info)
	return nil
}

// Pull gets a given chart from a chart registry and saves to specified location in tgz format.
// Returns the file name.
func (hc *HelmClient) Pull(chartRef string, outDir string) (string, error) {
	actionPull := action.NewPull()
	actionPull.DestDir = outDir
	actionPull.Settings = cli.New()

	output, err := actionPull.Run(chartRef)
	return output, err
}

func (hc *HelmClient) loadChart(chartRef string) (*chart.Chart, error) {
	chartPath := ""
	if strings.HasPrefix(chartRef, "file:") {
		urlRef, err := url.Parse(chartRef)
		if err != nil {
			logrus.Errorf("Error parsing file URL %v [%v]\n", chartRef, err.Error())
			return nil, err
		}
		chartPath = urlRef.RequestURI()
	} else {
		// If installing directly from repository, pull chart first and then install from temp filesystem location
		tmpDir, err := ioutil.TempDir(os.TempDir(), "helm-*")
		if err != nil {
			logrus.Errorf("Error creating temp directory %v %v", err.Error(), tmpDir)
			return nil, err
		}
		defer os.RemoveAll(tmpDir)
		chartPath, err = hc.Pull(chartRef, tmpDir)
		if err != nil {
			logrus.Errorf("Error pulling chart: %v", err.Error())
			return nil, err
		}
		f := findFilesInDir(tmpDir, ".tgz")
		if len(f) > 0 {
			chartPath = filepath.Join(tmpDir, f[0].Name())
		} else {
			return nil, errors.New("No chart file found in directory " + tmpDir)
		}
	}
	chart, err := loader.Load(chartPath)
	return chart, err
}

// Returns files with filename suffix
func findFilesInDir(directory string, suffix string) []os.FileInfo {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}
	outFiles := make([]os.FileInfo, 0)
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Name()), suffix) {
			outFiles = append(outFiles, file)
		}
	}
	return outFiles
}

// newHelmConfig returns the helm context for a given installation
func newHelmConfig(releaseNamespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	var settings = cli.New()
	err := actionConfig.Init(settings.RESTClientGetter(), releaseNamespace, os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		strContent := fmt.Sprintf(format, v)
		logrus.Infof("HELM: %v\n", strContent)
	})
	if err != nil {
		return nil, err
	}
	return actionConfig, err
}

// Load helm values files in the order specified by the array.  Later file entries will overwrite earlier ones.
func loadHelmValues(valueFiles []string) (map[string]interface{}, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		logrus.Infof("Reading values yaml file %v", filePath)
		bytes, err := readFile(filePath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return nil, err
		}
		// Merge with the previous map
		base = mergeMaps(base, currentMap)
	}
	base = convertStringKeyMap(base)
	return base, nil
}

// mergeMaps merges 2 maps returning the unified instance
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// convertStringKeyMap converts internal map[interface{}]interface{} types to map[string]interface{}
func convertStringKeyMap(m map[string]interface{}) map[string]interface{} {
	a := make(map[string]interface{}, len(m))
	logrus.Debugf("=======================================================")
	logrus.Debugf("map = %#v", m)
	for k, v := range m {
		if aMap, ok := v.(map[interface{}]interface{}); ok {
			logrus.Debugf("######### key=[%v]  valueMAP=[%v]", k, aMap)
			nm := make(map[string]interface{}, len(aMap))
			for kk, vv := range aMap {
				strKey := kk.(string)
				nm[strKey] = vv
			}
			a[k] = convertStringKeyMap(nm)
		} else {
			logrus.Debugf("######### key=[%v]  valueSTR=[%v]", k, v)
			a[k] = v
		}
	}
	return a
}

// readFile load a file from stdin, the local directory, or a remote file with a url.
func readFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}
