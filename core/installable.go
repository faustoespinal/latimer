package core

import "latimer/kube"

// Installable is an interface for any artifact which can be installed onto a system
type Installable interface {
	// Return string representation of the installable
	String() string

	// Return the yaml string representation of the installable
	StringYaml() string

	// Install the contents of the installable
	Install(sc *SystemContext) bool

	// Uninstall the contents of this installable
	Uninstall(sc *SystemContext) bool

	// Status returns the status of the installation within the given system context
	Status(sc *SystemContext) kube.InstallStatus

	// GetID returns the identifier name for this Installable.
	GetID() string
}
