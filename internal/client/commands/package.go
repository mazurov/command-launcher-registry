package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/criteo/command-launcher-registry/internal/client/errors"
	"github.com/criteo/command-launcher-registry/internal/client/output"
	"github.com/criteo/command-launcher-registry/internal/client/prompts"
	"github.com/criteo/command-launcher-registry/internal/client/validation"
	"github.com/spf13/cobra"
)

var (
	// Package command flags
	pkgDescription    string
	pkgMaintainers    []string
	pkgCustomValues   []string
	pkgClearMaint     bool
	pkgClearCustomVal bool
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage packages",
	Long:  `Create, list, get, update, and delete packages within a registry.`,
}

var packageCreateCmd = &cobra.Command{
	Use:   "create <registry> <package>",
	Short: "Create a new package",
	Args:  cobra.ExactArgs(2),
	Run:   runPackageCreate,
}

var packageListCmd = &cobra.Command{
	Use:   "list <registry>",
	Short: "List all packages in a registry",
	Args:  cobra.ExactArgs(1),
	Run:   runPackageList,
}

var packageGetCmd = &cobra.Command{
	Use:   "get <registry> <package>",
	Short: "Get package details",
	Args:  cobra.ExactArgs(2),
	Run:   runPackageGet,
}

var packageUpdateCmd = &cobra.Command{
	Use:   "update <registry> <package>",
	Short: "Update a package",
	Args:  cobra.ExactArgs(2),
	Run:   runPackageUpdate,
}

var packageDeleteCmd = &cobra.Command{
	Use:   "delete <registry> <package>",
	Short: "Delete a package",
	Args:  cobra.ExactArgs(2),
	Run:   runPackageDelete,
}

func init() {
	// Add subcommands
	packageCmd.AddCommand(packageCreateCmd)
	packageCmd.AddCommand(packageListCmd)
	packageCmd.AddCommand(packageGetCmd)
	packageCmd.AddCommand(packageUpdateCmd)
	packageCmd.AddCommand(packageDeleteCmd)

	// Create flags
	packageCreateCmd.Flags().StringVar(&pkgDescription, "description", "", "Package description")
	packageCreateCmd.Flags().StringSliceVar(&pkgMaintainers, "maintainer", []string{}, "Maintainer email (repeatable)")
	packageCreateCmd.Flags().StringSliceVar(&pkgCustomValues, "custom-value", []string{}, "Custom key=value (repeatable)")

	// Update flags
	packageUpdateCmd.Flags().StringVar(&pkgDescription, "description", "", "Package description")
	packageUpdateCmd.Flags().StringSliceVar(&pkgMaintainers, "maintainer", []string{}, "Maintainer email (repeatable, replaces all)")
	packageUpdateCmd.Flags().StringSliceVar(&pkgCustomValues, "custom-value", []string{}, "Custom key=value (repeatable, replaces all)")
	packageUpdateCmd.Flags().BoolVar(&pkgClearMaint, "clear-maintainers", false, "Clear all maintainers")
	packageUpdateCmd.Flags().BoolVar(&pkgClearCustomVal, "clear-custom-values", false, "Clear all custom values")

	rootCmd.AddCommand(packageCmd)
}

func runPackageCreate(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	c := getAuthenticatedClient()

	// Validate and parse custom values
	customValues, err := validation.ParseCustomValues(pkgCustomValues)
	if err != nil {
		errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
	}

	// Build request
	reqBody := map[string]interface{}{
		"name": packageName,
	}
	if pkgDescription != "" {
		reqBody["description"] = pkgDescription
	}
	if len(pkgMaintainers) > 0 {
		reqBody["maintainers"] = pkgMaintainers
	}
	if len(customValues) > 0 {
		reqBody["custom_values"] = customValues
	}

	resp, err := c.Post(fmt.Sprintf("/api/v1/registry/%s/package", registryName), reqBody)
	if err != nil {
		errors.ExitWithError(err, "failed to create package")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to create package: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]string{"registry": registryName, "package": packageName}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Created package '%s' in registry '%s'", packageName, registryName))
	}
}

func runPackageList(cmd *cobra.Command, args []string) {
	registryName := args[0]
	c := getAuthenticatedClient()

	resp, err := c.Get(fmt.Sprintf("/api/v1/registry/%s/package", registryName))
	if err != nil {
		errors.ExitWithError(err, "failed to list packages")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to list packages: %s", string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ExitWithError(err, "failed to read response")
	}

	var packages []map[string]interface{}
	if err := json.Unmarshal(body, &packages); err != nil {
		errors.ExitWithError(err, "failed to parse response")
	}

	if flagJSON {
		output.OutputJSON(packages, nil)
	} else {
		if len(packages) == 0 {
			fmt.Printf("No packages found in registry '%s'\n", registryName)
			return
		}

		table := output.NewTableWriter()
		table.WriteHeader("NAME", "DESCRIPTION", "VERSIONS")
		for _, pkg := range packages {
			name := fmt.Sprintf("%v", pkg["name"])
			description := fmt.Sprintf("%v", pkg["description"])
			versions := "0"
			// Versions are returned as a map, not array
			if vers, ok := pkg["versions"].(map[string]interface{}); ok {
				versions = strconv.Itoa(len(vers))
			}
			table.WriteRow(name, description, versions)
		}
		table.Flush()
	}
}

func runPackageGet(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	c := getAuthenticatedClient()

	resp, err := c.Get(fmt.Sprintf("/api/v1/registry/%s/package/%s", registryName, packageName))
	if err != nil {
		errors.ExitWithError(err, "failed to get package")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to get package: %s", string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ExitWithError(err, "failed to read response")
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(body, &pkg); err != nil {
		errors.ExitWithError(err, "failed to parse response")
	}

	if flagJSON {
		output.OutputJSON(pkg, nil)
	} else {
		fmt.Printf("Name: %v\n", pkg["name"])
		fmt.Printf("Description: %v\n", pkg["description"])
		if maintainers, ok := pkg["maintainers"].([]interface{}); ok && len(maintainers) > 0 {
			fmt.Print("Maintainers:")
			for _, maintainer := range maintainers {
				fmt.Printf("\n  - %v", maintainer)
			}
			fmt.Println()
		}
		if customVals, ok := pkg["custom_values"].(map[string]interface{}); ok && len(customVals) > 0 {
			fmt.Println("Custom Values:")
			for k, v := range customVals {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
		if versions, ok := pkg["versions"].([]interface{}); ok && len(versions) > 0 {
			fmt.Printf("Versions: %d\n", len(versions))
		}
	}
}

func runPackageUpdate(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	c := getAuthenticatedClient()

	// Validate flag conflicts
	if pkgClearMaint && len(pkgMaintainers) > 0 {
		errors.ExitWithCode(errors.ExitInvalidArguments, "cannot use --clear-maintainers with --maintainer. Use one or the other")
	}
	if pkgClearCustomVal && len(pkgCustomValues) > 0 {
		errors.ExitWithCode(errors.ExitInvalidArguments, "cannot use --clear-custom-values with --custom-value. Use one or the other")
	}

	// Validate and parse custom values
	var customValues map[string]string
	if len(pkgCustomValues) > 0 {
		var err error
		customValues, err = validation.ParseCustomValues(pkgCustomValues)
		if err != nil {
			errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
		}
	}

	// Build partial update request
	reqBody := make(map[string]interface{})
	if pkgDescription != "" {
		reqBody["description"] = pkgDescription
	}
	if pkgClearMaint {
		reqBody["maintainers"] = []string{}
	} else if len(pkgMaintainers) > 0 {
		reqBody["maintainers"] = pkgMaintainers
	}
	if pkgClearCustomVal {
		reqBody["custom_values"] = map[string]string{}
	} else if len(customValues) > 0 {
		reqBody["custom_values"] = customValues
	}

	resp, err := c.Put(fmt.Sprintf("/api/v1/registry/%s/package/%s", registryName, packageName), reqBody)
	if err != nil {
		errors.ExitWithError(err, "failed to update package")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to update package: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]string{"registry": registryName, "package": packageName}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Updated package '%s' in registry '%s'", packageName, registryName))
	}
}

func runPackageDelete(cmd *cobra.Command, args []string) {
	registryName := args[0]
	packageName := args[1]
	c := getAuthenticatedClient()

	// Prompt for confirmation unless --yes flag is set
	if !flagYes {
		if !prompts.ConfirmDeletion("package", packageName, "all its versions") {
			fmt.Println("Deletion cancelled")
			return
		}
	}

	resp, err := c.Delete(fmt.Sprintf("/api/v1/registry/%s/package/%s", registryName, packageName))
	if err != nil {
		errors.ExitWithError(err, "failed to delete package")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to delete package: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]bool{"deleted": true}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Deleted package '%s' from registry '%s'", packageName, registryName))
	}
}
