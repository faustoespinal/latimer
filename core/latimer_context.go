package core

import (
	"io/ioutil"
	"latimer/kube"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LatimerContext contains the values derived from global command arguments
type LatimerContext struct {
	KubeConfigPath string
	ManifestPath   string
	KubeClient     *kube.K8sClient
	LatimerTempDir string
	Values         map[string]string
}

var lc *LatimerContext = nil

// GetLatimerContext creates a LatimerContext object
func GetLatimerContext() *LatimerContext {
	if lc == nil {
		logrus.Debugf("Creating LatimerContext\n")
		lc = new(LatimerContext)
		lc.Values = map[string]string{}
	}
	return lc
}

// InitLatimer loads the kube config settings and initialize basic settings
func (latimerContext *LatimerContext) InitLatimer(kubeConfigPath string, manifestPath string, values []string) {
	var err error
	latimerContext.KubeConfigPath = kubeConfigPath
	for _, valueItem := range values {
		keyvals := strings.Split(valueItem, ",")
		for _, keyvalItem := range keyvals {
			keyval := strings.Split(keyvalItem, "=")
			if len(keyval) >= 2 {
				key := strings.TrimSpace(keyval[0])
				value := strings.TrimSpace(keyval[1])
				latimerContext.Values[key] = value
			}
		}
	}
	logrus.Infof("LATIMER VALUES=[%v]", latimerContext.Values)
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
	logrus.Infof("INITIALIZED LATIMER CONTEXT kubeConfig=%v", latimerContext.KubeConfigPath)
}
