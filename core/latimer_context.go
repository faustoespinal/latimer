package core

import (
	"io/ioutil"
	"latimer/kube"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

// LatimerContext contains the values derived from global command arguments
type LatimerContext struct {
	KubeConfigPath string
	ManifestPath   string
	KubeClient     *kube.K8sClient
	LatimerTempDir string
	ChartRegistry  string
}

var lc *LatimerContext = nil

// GetLatimerContext creates a LatimerContext object
func GetLatimerContext() *LatimerContext {
	if lc == nil {
		logrus.Debugf("Creating LatimerContext\n")
		lc = new(LatimerContext)
	}
	return lc
}

// InitLatimer loads the kube config settings and initialize basic settings
func (latimerContext *LatimerContext) InitLatimer(kubeConfigPath string, manifestPath string, chartRegistry string) {
	var err error
	latimerContext.KubeConfigPath = kubeConfigPath
	latimerContext.ChartRegistry = chartRegistry
	latimerContext.KubeClient, err = kube.NewK8sClient(kubeConfigPath)
	if err != nil {
		panic(err.Error())
	}
	latimerContext.ManifestPath = manifestPath
	tmpDir, err := ioutil.TempDir(os.TempDir(), "*-latimer")
	if err != nil {
		log.Fatal(err)
	}
	latimerContext.LatimerTempDir = tmpDir
	logrus.Infof("INITIALIZED LATIMER CONTEXT kubeConfig=%v chartRegistry=%v", latimerContext.KubeConfigPath, latimerContext.ChartRegistry)
}
