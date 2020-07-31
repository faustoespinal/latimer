package core

import (
	"net/url"
)

// SystemContext class describing an install target
type SystemContext struct {
	// Name of the system context
	Name string `json:"name"`

	// Address of the system to which an Installable can target.
	URL url.URL `json:"systemAddress"`

	// ReleaseName is the name of the installed instance (e.g. release name in a tool like HELM)
	ReleaseName string `json:"releaseName"`

	// The deployment space where to deploy the Installable
	DeploymentSpace string `json:"deploymentSpace"`

	// Work temporary directory to be used for temporary files in an install or uninstall
	WorkTempDir string

	// Context is the global context info
	Context *LatimerContext
}
