package core

// InstallableItem represents a name identifier/type pair
type InstallableItem struct {
	// Name is the name of the item
	Name string `json:"name"`
	// Kind is the type of the item
	Kind string `json:"kind"`
}
