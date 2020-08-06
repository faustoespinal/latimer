/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io/ioutil"
	"latimer/core"
	"latimer/manifest"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs a collection of charts and packages defined in a manifest file input",
	Long: `Installs a collection of charts and packages defined in a manifest file input.
A sample manifest file looks like:

metadata:
  name: install-manifest
  kind: manifest
charts:
  - name: "redis"
    chartName: "bitnami/redis"
    namespace: "paas"
    chartLocator: "{{.ChartLocation}}/redis-10.7.9.tgz"
    releaseName: "test-redis"
    values:
      - url: "{{.ChartLocation}}/redis/values.yaml"`,
	Run: func(cmd *cobra.Command, args []string) {
		latimerContext := core.GetLatimerContext()
		filePath := latimerContext.ManifestPath
		logrus.Infof("Install %v\n", filePath)
		manifest, err := manifest.NewManifest(filePath, latimerContext.Values)
		if err != nil {
			logrus.Errorf("Error loading manifest file: %v", filePath)
			os.Exit(1)
		}
		logrus.Infof("\n%v\n", manifest.StringYaml())

		descriptor := manifest.Descriptor
		// Each installable should work in it's own private temp directory
		installableTempDir, err := ioutil.TempDir(latimerContext.LatimerTempDir, descriptor.Metadata.Name+"-*")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(installableTempDir) // clean up

		sc := &core.SystemContext{
			Name:        descriptor.Metadata.Name,
			WorkTempDir: installableTempDir,
			Context:     latimerContext,
		}
		manifest.Install(sc)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
