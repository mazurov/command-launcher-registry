package storage

import (
	"fmt"
	"net/url"
	"strings"
)

// SupportedSchemes lists all currently supported storage URI schemes
var SupportedSchemes = []string{"file", "oci"}

// PlannedSchemes lists schemes that are recognized but not yet implemented
var PlannedSchemes = []string{}

// StorageURI represents a parsed storage backend URI
type StorageURI struct {
	Scheme string // Storage backend type (e.g., "file", "oci")
	Host   string // Host for network backends (optional for file://)
	Path   string // Path to storage resource
	Raw    string // Original URI string for logging/debugging
}

// NormalizeStorageURI ensures the URI has a scheme, prepending "file://" if missing
func NormalizeStorageURI(uri string) string {
	if uri == "" {
		return uri
	}
	if !strings.Contains(uri, "://") {
		return "file://" + uri
	}
	return uri
}

// ParseStorageURI parses a storage URI string into its components
func ParseStorageURI(uri string) (*StorageURI, error) {
	if uri == "" {
		return nil, fmt.Errorf("storage URI cannot be empty")
	}

	// Normalize URI (add file:// if no scheme)
	normalized := NormalizeStorageURI(uri)

	// Parse using net/url
	parsed, err := url.Parse(normalized)
	if err != nil {
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	// Validate scheme is present
	if parsed.Scheme == "" {
		return nil, fmt.Errorf("URI must have a scheme (e.g., file://)")
	}

	// Check if scheme is supported
	if err := validateScheme(parsed.Scheme); err != nil {
		return nil, err
	}

	// OCI-specific validation: reject query params and fragments (FR-015)
	if parsed.Scheme == "oci" {
		if parsed.RawQuery != "" {
			return nil, fmt.Errorf("OCI URI does not support query parameters")
		}
		if parsed.Fragment != "" {
			return nil, fmt.Errorf("OCI URI does not support fragments")
		}
		if parsed.Host == "" {
			return nil, fmt.Errorf("OCI URI must include registry host: oci://<registry>/<repository>")
		}
		// Remove leading slash from path for OCI
		ociPath := strings.TrimPrefix(parsed.Path, "/")
		if ociPath == "" {
			return nil, fmt.Errorf("OCI URI must include repository path: oci://<registry>/<repository>")
		}
		// Strip any tag from path (we always use :latest)
		if idx := strings.LastIndex(ociPath, ":"); idx > 0 {
			ociPath = ociPath[:idx]
		}
		return &StorageURI{
			Scheme: parsed.Scheme,
			Host:   parsed.Host,
			Path:   ociPath,
			Raw:    uri,
		}, nil
	}

	// Extract path - for file:// URIs, the path may be in different places
	path := parsed.Path
	if parsed.Scheme == "file" {
		// Handle file:// URIs - path might be in Opaque for relative paths
		if path == "" && parsed.Opaque != "" {
			path = parsed.Opaque
		}
		// For file://./path format, the path starts with ./
		if parsed.Host == "." && strings.HasPrefix(path, "/") {
			path = "./" + strings.TrimPrefix(path, "/")
		} else if parsed.Host != "" && path != "" {
			// Windows-style: file:///C:/path or file://C:/path
			// Keep path as-is for Windows paths
			if len(parsed.Host) == 1 && strings.ToUpper(parsed.Host) >= "A" && strings.ToUpper(parsed.Host) <= "Z" {
				// Windows drive letter detected: file://C:/path
				path = parsed.Host + ":" + path
			}
		}
	}

	if path == "" {
		return nil, fmt.Errorf("storage URI must have a path")
	}

	return &StorageURI{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   path,
		Raw:    uri,
	}, nil
}

// validateScheme checks if the scheme is supported or planned
func validateScheme(scheme string) error {
	// Check supported schemes
	for _, s := range SupportedSchemes {
		if scheme == s {
			return nil
		}
	}

	// Check planned but not implemented schemes
	for _, s := range PlannedSchemes {
		if scheme == s {
			return fmt.Errorf("storage scheme %q is not yet implemented (planned for future release); supported schemes: %s",
				scheme, strings.Join(SupportedSchemes, ", "))
		}
	}

	// Unknown scheme
	return fmt.Errorf("unsupported storage scheme %q; supported schemes: %s",
		scheme, strings.Join(SupportedSchemes, ", "))
}

// IsFileScheme returns true if this is a file:// URI
func (u *StorageURI) IsFileScheme() bool {
	return u.Scheme == "file"
}

// IsOCIScheme returns true if this is an oci:// URI
func (u *StorageURI) IsOCIScheme() bool {
	return u.Scheme == "oci"
}

// OCIReference returns the OCI reference string "registry/repository:latest"
// This should only be called for OCI scheme URIs
func (u *StorageURI) OCIReference() string {
	return fmt.Sprintf("%s/%s:latest", u.Host, u.Path)
}

// String returns the original URI string
func (u *StorageURI) String() string {
	return u.Raw
}
