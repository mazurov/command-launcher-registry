package storage

import (
	"fmt"
	"net/url"
	"strings"
)

// SupportedSchemes lists all currently supported storage URI schemes
var SupportedSchemes = []string{"file"}

// PlannedSchemes lists schemes that are recognized but not yet implemented
var PlannedSchemes = []string{"oci"}

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

// String returns the original URI string
func (u *StorageURI) String() string {
	return u.Raw
}
