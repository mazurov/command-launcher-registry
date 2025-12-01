package validation

import (
	"fmt"
	"strings"
)

// ValidateChecksum validates checksum format (must start with "sha256:")
func ValidateChecksum(checksum string) error {
	if !strings.HasPrefix(checksum, "sha256:") {
		return fmt.Errorf("invalid checksum format. Expected 'sha256:hash', got: '%s'", checksum)
	}
	return nil
}

// ValidatePartitionRange validates partition range (0-9)
func ValidatePartitionRange(start, end int) error {
	if start < 0 || start > 9 {
		return fmt.Errorf("invalid start partition. Must be 0-9, got: %d", start)
	}
	if end < 0 || end > 9 {
		return fmt.Errorf("invalid end partition. Must be 0-9, got: %d", end)
	}
	if start > end {
		return fmt.Errorf("start partition (%d) cannot be greater than end partition (%d)", start, end)
	}
	return nil
}

// ValidateCustomValue validates custom value format (key=value)
func ValidateCustomValue(customValue string) (key, value string, err error) {
	parts := strings.SplitN(customValue, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid --custom-value format. Expected 'key=value', got: '%s'", customValue)
	}

	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", fmt.Errorf("invalid --custom-value format. Key cannot be empty in '%s'", customValue)
	}
	if value == "" {
		return "", "", fmt.Errorf("invalid --custom-value format. Value cannot be empty in '%s'", customValue)
	}

	return key, value, nil
}

// ParseCustomValues parses a slice of custom values into a map
func ParseCustomValues(customValues []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, cv := range customValues {
		key, value, err := ValidateCustomValue(cv)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}
