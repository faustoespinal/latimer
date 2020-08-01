/*
Copyright © 2020 Fausto J Espinal

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"io"
	"latimer/core"
	"os"
	"os/user"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var kubeConfigPath string
var manifestPath string
var valuesLatimer []string = []string{}

//The verbose flag value
var verbosity string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "latimer",
	Short: "Latimer is a k8s package installation orchestration tool",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	latimerContext := core.GetLatimerContext()
	defer os.RemoveAll(latimerContext.LatimerTempDir) // clean up
}

func init() {
	cobra.OnInitialize(initConfig, initLatimer)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setUpLogs(os.Stdout, verbosity); err != nil {
			return err
		}
		return nil
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	defaultKubeConfigPath := filepath.Join(user.HomeDir, ".kube", "config")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.latimer.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", defaultKubeConfigPath, "kubeconfig file (default is $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&manifestPath, "manifest", "default", "Path of the input manifest")
	//Default value is the warn level
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", logrus.WarnLevel.String(), "Log level (debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringArrayVar(&valuesLatimer, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".latimer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".latimer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Load the kube config settings
func initLatimer() {
	latimerContext := core.GetLatimerContext()
	latimerContext.InitLatimer(kubeConfigPath, manifestPath, valuesLatimer)
}

//setUpLogs set the log output ans the log level
func setUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}
