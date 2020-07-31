package core

// InstallStatus models the different status codes for an installable
type InstallStatus int

const (
	// NotReady means the Installable is in process of installation/uninstallation but not ready
	NotReady InstallStatus = iota
	// Ready means Installable is fully installed and operational
	Ready
	// Uninstalled means the Installable has been uninstalled fully
	Uninstalled
)

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

	// Status returns the status of the  installation
	Status() InstallStatus

	// GetID returns the identifier name for this Installable.
	GetID() string
}
