package types

// PackageInfo represents package version information (compatible with CDT client)
type PackageInfo struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	URL            string `json:"url"`
	Checksum       string `json:"checksum"`
	StartPartition uint8  `json:"startPartition"`
	EndPartition   uint8  `json:"endPartition"`
}

// RegistryIndex represents the CDT-compatible registry index format
type RegistryIndex struct {
	Packages map[string]PackageVersions `json:"packages"`
}

// PackageVersions represents all versions of a package
type PackageVersions struct {
	Versions []PackageInfo `json:"versions"`
}

// CreateRegistryRequest represents the request to create a registry
type CreateRegistryRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Admin        []string          `json:"admin"`
	CustomValues map[string]string `json:"customValues"`
}

// UpdateRegistryRequest represents the request to update a registry
type UpdateRegistryRequest struct {
	Description  string            `json:"description"`
	Admin        []string          `json:"admin"`
	CustomValues map[string]string `json:"customValues"`
}

// CreatePackageRequest represents the request to create a package
type CreatePackageRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Admin        []string          `json:"admin"`
	CustomValues map[string]string `json:"customValues"`
}

// UpdatePackageRequest represents the request to update a package
type UpdatePackageRequest struct {
	Description  string            `json:"description"`
	Admin        []string          `json:"admin"`
	CustomValues map[string]string `json:"customValues"`
}

// PublishVersionRequest represents the request to publish a new version
type PublishVersionRequest struct {
	Version        string `json:"version" binding:"required"`
	URL            string `json:"url" binding:"required"`
	Checksum       string `json:"checksum"`
	StartPartition uint8  `json:"startPartition"`
	EndPartition   uint8  `json:"endPartition"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
