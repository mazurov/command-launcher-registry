package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/criteo/command-launcher-registry/internal/client/errors"
	"github.com/criteo/command-launcher-registry/internal/client/output"
	"github.com/criteo/command-launcher-registry/internal/client/prompts"
	"github.com/spf13/cobra"
)

var (
	// Version command flags
	versionChecksum     string
	versionURL          string
	versionStartPart    int
	versionEndPart      int
	versionStartPartSet bool
	versionEndPartSet   bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage package versions",
	Long:  `Create, list, get, and delete package versions.`,
}

var versionCreateCmd = &cobra.Command{
	Use:   "create <registry> <package> <version>",
	Short: "Create a new version",
	Args:  cobra.ExactArgs(3),
	Run:   runVersionCreate,
}

var versionListCmd = &cobra.Command{
	Use:   "list <registry> <package>",
	Short: "List all versions of a package",
	Args:  cobra.ExactArgs(2),
	Run:   runVersionList,
}

var versionGetCmd = &cobra.Command{
	Use:   "get <registry> <package> <version>",
	Short: "Get version details",
	Args:  cobra.ExactArgs(3),
	Run:   runVersionGet,
}

var versionDeleteCmd = &cobra.Command{
	Use:   "delete <registry> <package> <version>",
	Short: "Delete a version",
	Args:  cobra.ExactArgs(3),
	Run:   runVersionDelete,
}

func init() {
	// Add subcommands
	versionCmd.AddCommand(versionCreateCmd)
	versionCmd.AddCommand(versionListCmd)
	versionCmd.AddCommand(versionGetCmd)
	versionCmd.AddCommand(versionDeleteCmd)

	// Create flags
	versionCreateCmd.Flags().StringVar(&versionChecksum, "checksum", "", "Checksum in format 'sha256:hash' (required)")
	versionCreateCmd.Flags().StringVar(&versionURL, "url", "", "Download URL (required)")
	versionCreateCmd.Flags().IntVar(&versionStartPart, "start-partition", 0, "Start partition (0-9)")
	versionCreateCmd.Flags().IntVar(&versionEndPart, "end-partition", 9, "End partition (0-9)")

	// Mark required flags
	versionCreateCmd.MarkFlagRequired("checksum")
	versionCreateCmd.MarkFlagRequired("url")

	rootCmd.AddCommand(versionCmd)
}

func validateChecksum(checksum string) error {
	if !strings.HasPrefix(checksum, "sha256:") {
		return fmt.Errorf("checksum must start with 'sha256:'")
	}

	hash := strings.TrimPrefix(checksum, "sha256:")
	if len(hash) != 64 {
		return fmt.Errorf("sha256 hash must be exactly 64 hexadecimal characters")
	}

	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("sha256 hash must contain only hexadecimal characters")
		}
	}

	return nil
}

func validatePartitionRange(start, end int) error {
	if start < 0 || start > 9 {
		return fmt.Errorf("start partition must be between 0 and 9")
	}
	if end < 0 || end > 9 {
		return fmt.Errorf("end partition must be between 0 and 9")
	}
	if start > end {
		return fmt.Errorf("start partition (%d) cannot be greater than end partition (%d)", start, end)
	}
	return nil
}

func runVersionCreate(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	versionName := args[2]
	c := getAuthenticatedClient()

	// Validate checksum format
	if err := validateChecksum(versionChecksum); err != nil {
		errors.ExitWithCode(errors.ExitInvalidArguments, fmt.Sprintf("invalid checksum: %s", err.Error()))
	}

	// Validate partition range
	if err := validatePartitionRange(versionStartPart, versionEndPart); err != nil {
		errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
	}

	// Build request
	reqBody := map[string]interface{}{
		"name":           packageName,
		"version":        versionName,
		"checksum":       versionChecksum,
		"url":            versionURL,
		"startPartition": versionStartPart,
		"endPartition":   versionEndPart,
	}

	resp, err := c.Post(fmt.Sprintf("/api/v1/registry/%s/package/%s/version", registryName, packageName), reqBody)
	if err != nil {
		errors.ExitWithError(err, "failed to create version")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to create version: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]string{
			"registry": registryName,
			"package":  packageName,
			"version":  versionName,
		}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Created version '%s' for package '%s' in registry '%s'", versionName, packageName, registryName))
	}
}

func runVersionList(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	c := getAuthenticatedClient()

	resp, err := c.Get(fmt.Sprintf("/api/v1/registry/%s/package/%s/version", registryName, packageName))
	if err != nil {
		errors.ExitWithError(err, "failed to list versions")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to list versions: %s", string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ExitWithError(err, "failed to read response")
	}

	var versions []map[string]interface{}
	if err := json.Unmarshal(body, &versions); err != nil {
		errors.ExitWithError(err, "failed to parse response")
	}

	if flagJSON {
		output.OutputJSON(versions, nil)
	} else {
		if len(versions) == 0 {
			fmt.Printf("No versions found for package '%s' in registry '%s'\n", packageName, registryName)
			return
		}

		table := output.NewTableWriter()
		table.WriteHeader("VERSION", "CHECKSUM", "PARTITIONS")
		for _, ver := range versions {
			version := fmt.Sprintf("%v", ver["version"])
			checksum := fmt.Sprintf("%v", ver["checksum"])
			if len(checksum) > 20 {
				checksum = checksum[:17] + "..."
			}

			startPart := 0
			endPart := 9
			if sp, ok := ver["startPartition"].(float64); ok {
				startPart = int(sp)
			}
			if ep, ok := ver["endPartition"].(float64); ok {
				endPart = int(ep)
			}
			partitions := fmt.Sprintf("%d-%d", startPart, endPart)

			table.WriteRow(version, checksum, partitions)
		}
		table.Flush()
	}
}

func runVersionGet(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	versionName := args[2]
	c := getAuthenticatedClient()

	resp, err := c.Get(fmt.Sprintf("/api/v1/registry/%s/package/%s/version/%s", registryName, packageName, versionName))
	if err != nil {
		errors.ExitWithError(err, "failed to get version")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to get version: %s", string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ExitWithError(err, "failed to read response")
	}

	var version map[string]interface{}
	if err := json.Unmarshal(body, &version); err != nil {
		errors.ExitWithError(err, "failed to parse response")
	}

	if flagJSON {
		output.OutputJSON(version, nil)
	} else {
		fmt.Printf("Version: %v\n", version["version"])
		fmt.Printf("Checksum: %v\n", version["checksum"])
		fmt.Printf("URL: %v\n", version["url"])

		startPart := 0
		endPart := 9
		if sp, ok := version["startPartition"].(float64); ok {
			startPart = int(sp)
		}
		if ep, ok := version["endPartition"].(float64); ok {
			endPart = int(ep)
		}
		fmt.Printf("Partition Range: %d-%d\n", startPart, endPart)
	}
}

func runVersionDelete(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	versionName := args[2]
	c := getAuthenticatedClient()

	// Prompt for confirmation unless --yes flag is set
	if !flagYes {
		if !prompts.ConfirmDeletion("version", versionName, "") {
			fmt.Println("Deletion cancelled")
			return
		}
	}

	resp, err := c.Delete(fmt.Sprintf("/api/v1/registry/%s/package/%s/version/%s", registryName, packageName, versionName))
	if err != nil {
		errors.ExitWithError(err, "failed to delete version")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to delete version: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]bool{"deleted": true}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Deleted version '%s' from package '%s' in registry '%s'", versionName, packageName, registryName))
	}
}
