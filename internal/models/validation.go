package models

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	// Name pattern: 1-64 characters, alphanumeric with hyphens/underscores
	namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

	// Semantic version pattern (simplified - supports major.minor.patch with optional pre-release and build metadata)
	versionPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

	// Checksum pattern: sha256: followed by 64 hex characters
	checksumPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

	// Custom values key pattern
	customKeyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]{0,63}$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ValidateName validates registry or package name
func ValidateName(name string) error {
	if len(name) == 0 {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if len(name) > 64 {
		return &ValidationError{Field: "name", Message: "name must be at most 64 characters"}
	}
	if !namePattern.MatchString(name) {
		return &ValidationError{Field: "name", Message: "name must match pattern ^[a-z0-9][a-z0-9_-]*$"}
	}
	return nil
}

// ValidateDescription validates description field
func ValidateDescription(description string) error {
	if len(description) > 4096 {
		return &ValidationError{Field: "description", Message: "description must be at most 4096 characters"}
	}
	return nil
}

// ValidateVersion validates semantic version string
func ValidateVersion(version string) error {
	if len(version) == 0 {
		return &ValidationError{Field: "version", Message: "version is required"}
	}
	if !versionPattern.MatchString(version) {
		return &ValidationError{Field: "version", Message: "version must be valid semantic version (e.g., 1.0.0, 2.1.3-alpha)"}
	}
	return nil
}

// ValidateChecksum validates SHA256 checksum format
func ValidateChecksum(checksum string) error {
	if len(checksum) == 0 {
		return &ValidationError{Field: "checksum", Message: "checksum is required"}
	}
	if !checksumPattern.MatchString(checksum) {
		return &ValidationError{Field: "checksum", Message: "checksum must match format sha256:[64 hex characters]"}
	}
	return nil
}

// ValidateURL validates URL format (not reachability)
func ValidateURL(urlStr string) error {
	if len(urlStr) == 0 {
		return &ValidationError{Field: "url", Message: "url is required"}
	}
	if len(urlStr) > 2048 {
		return &ValidationError{Field: "url", Message: "url must be at most 2048 characters"}
	}

	// Parse URL to validate RFC 3986 syntax
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return &ValidationError{Field: "url", Message: fmt.Sprintf("url must be valid RFC 3986 URI: %v", err)}
	}

	// Must have http or https scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &ValidationError{Field: "url", Message: "url must start with http:// or https://"}
	}

	return nil
}

// ValidatePartitions validates partition range
func ValidatePartitions(startPartition, endPartition int) error {
	if startPartition < 0 || startPartition > 9 {
		return &ValidationError{Field: "startPartition", Message: "startPartition must be in range 0-9"}
	}
	if endPartition < 0 || endPartition > 9 {
		return &ValidationError{Field: "endPartition", Message: "endPartition must be in range 0-9"}
	}
	if startPartition > endPartition {
		return &ValidationError{Field: "partitions", Message: "startPartition must be <= endPartition"}
	}
	return nil
}

// CheckPartitionOverlap checks if two partition ranges overlap
func CheckPartitionOverlap(start1, end1, start2, end2 int) bool {
	// Ranges overlap if: start1 <= end2 && start2 <= end1
	return start1 <= end2 && start2 <= end1
}

// ValidateCustomValues validates custom_values map
func ValidateCustomValues(customValues map[string]string) error {
	if len(customValues) > 20 {
		return &ValidationError{
			Field:   "custom_values",
			Message: "custom_values must contain at most 20 key-value pairs",
		}
	}

	for key, value := range customValues {
		// Validate key pattern
		if len(key) == 0 || len(key) > 64 {
			return &ValidationError{
				Field:   "custom_values",
				Message: fmt.Sprintf("custom_values key '%s' must be 1-64 characters", key),
			}
		}
		if !customKeyPattern.MatchString(key) {
			return &ValidationError{
				Field:   "custom_values",
				Message: fmt.Sprintf("custom_values key '%s' must match pattern ^[a-zA-Z_][a-zA-Z0-9_-]{0,63}$", key),
			}
		}

		// Validate value length
		if len(value) > 1024 {
			return &ValidationError{
				Field:   "custom_values",
				Message: fmt.Sprintf("custom_values value for key '%s' must be at most 1024 characters", key),
			}
		}
	}

	return nil
}

// ValidateRegistry validates a registry
func ValidateRegistry(r *Registry) error {
	if err := ValidateName(r.Name); err != nil {
		return err
	}
	if err := ValidateDescription(r.Description); err != nil {
		return err
	}
	if err := ValidateCustomValues(r.CustomValues); err != nil {
		return err
	}
	return nil
}

// ValidatePackage validates a package
func ValidatePackage(p *Package) error {
	if err := ValidateName(p.Name); err != nil {
		return err
	}
	if err := ValidateDescription(p.Description); err != nil {
		return err
	}
	if err := ValidateCustomValues(p.CustomValues); err != nil {
		return err
	}
	return nil
}

// ValidateVersionData validates version data
func ValidateVersionData(v *Version) error {
	if err := ValidateVersion(v.Version); err != nil {
		return err
	}
	if err := ValidateChecksum(v.Checksum); err != nil {
		return err
	}
	if err := ValidateURL(v.URL); err != nil {
		return err
	}
	if err := ValidatePartitions(v.StartPartition, v.EndPartition); err != nil {
		return err
	}
	return nil
}

// NormalizeName converts name to lowercase for case-insensitive comparison
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
