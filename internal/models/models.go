package models

// Registry represents a named container for packages
type Registry struct {
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Admins       []string            `json:"admins,omitempty"`
	CustomValues map[string]string   `json:"custom_values,omitempty"`
	Packages     map[string]*Package `json:"packages"`
}

// Package represents metadata for a command bundle within a registry
type Package struct {
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Maintainers  []string            `json:"maintainers,omitempty"`
	CustomValues map[string]string   `json:"custom_values,omitempty"`
	Versions     map[string]*Version `json:"versions"`
}

// Version represents a specific release of a package (immutable)
type Version struct {
	Name           string `json:"name"` // Package name (denormalized for index.json)
	Version        string `json:"version"`
	Checksum       string `json:"checksum"`       // SHA256 with "sha256:" prefix
	URL            string `json:"url"`            // Download URL
	StartPartition int    `json:"startPartition"` // 0-9
	EndPartition   int    `json:"endPartition"`   // 0-9
}

// IndexEntry represents an entry in the registry index.json (Command Launcher format)
type IndexEntry struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Checksum       string `json:"checksum"`
	URL            string `json:"url"`
	StartPartition int    `json:"startPartition"`
	EndPartition   int    `json:"endPartition"`
}

// Storage is the root storage structure
type Storage struct {
	Registries map[string]*Registry `json:"registries"`
}

// NewStorage creates an empty storage structure
func NewStorage() *Storage {
	return &Storage{
		Registries: make(map[string]*Registry),
	}
}

// NewRegistry creates a new registry with initialized maps
func NewRegistry(name, description string, admins []string, customValues map[string]string) *Registry {
	if customValues == nil {
		customValues = make(map[string]string)
	}
	return &Registry{
		Name:         name,
		Description:  description,
		Admins:       admins,
		CustomValues: customValues,
		Packages:     make(map[string]*Package),
	}
}

// NewPackage creates a new package with initialized maps
func NewPackage(name, description string, maintainers []string, customValues map[string]string) *Package {
	if customValues == nil {
		customValues = make(map[string]string)
	}
	return &Package{
		Name:         name,
		Description:  description,
		Maintainers:  maintainers,
		CustomValues: customValues,
		Versions:     make(map[string]*Version),
	}
}

// NewVersion creates a new version
func NewVersion(name, version, checksum, url string, startPartition, endPartition int) *Version {
	return &Version{
		Name:           name,
		Version:        version,
		Checksum:       checksum,
		URL:            url,
		StartPartition: startPartition,
		EndPartition:   endPartition,
	}
}

// ToIndexEntry converts a Version to an IndexEntry
func (v *Version) ToIndexEntry() IndexEntry {
	return IndexEntry{
		Name:           v.Name,
		Version:        v.Version,
		Checksum:       v.Checksum,
		URL:            v.URL,
		StartPartition: v.StartPartition,
		EndPartition:   v.EndPartition,
	}
}
