package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/criteo/command-launcher-registry/internal/client"
	"github.com/criteo/command-launcher-registry/internal/client/auth"
	"github.com/criteo/command-launcher-registry/internal/client/config"
	"github.com/criteo/command-launcher-registry/internal/client/errors"
	"github.com/criteo/command-launcher-registry/internal/client/output"
	"github.com/criteo/command-launcher-registry/internal/client/prompts"
	"github.com/criteo/command-launcher-registry/internal/client/validation"
	"github.com/spf13/cobra"
)

var (
	// Registry command flags
	regDescription    string
	regAdmins         []string
	regCustomValues   []string
	regClearAdmins    bool
	regClearCustomVal bool
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage registries",
	Long:  `Create, list, get, update, and delete registries.`,
}

var registryCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new registry",
	Args:  cobra.ExactArgs(1),
	Run:   runRegistryCreate,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registries",
	Args:  cobra.NoArgs,
	Run:   runRegistryList,
}

var registryGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get registry details",
	Args:  cobra.ExactArgs(1),
	Run:   runRegistryGet,
}

var registryUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a registry",
	Args:  cobra.ExactArgs(1),
	Run:   runRegistryUpdate,
}

var registryDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a registry",
	Args:  cobra.ExactArgs(1),
	Run:   runRegistryDelete,
}

func init() {
	// Add subcommands
	registryCmd.AddCommand(registryCreateCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryGetCmd)
	registryCmd.AddCommand(registryUpdateCmd)
	registryCmd.AddCommand(registryDeleteCmd)

	// Create flags
	registryCreateCmd.Flags().StringVar(&regDescription, "description", "", "Registry description")
	registryCreateCmd.Flags().StringSliceVar(&regAdmins, "admin", []string{}, "Admin email (repeatable)")
	registryCreateCmd.Flags().StringSliceVar(&regCustomValues, "custom-value", []string{}, "Custom key=value (repeatable)")

	// Update flags
	registryUpdateCmd.Flags().StringVar(&regDescription, "description", "", "Registry description")
	registryUpdateCmd.Flags().StringSliceVar(&regAdmins, "admin", []string{}, "Admin email (repeatable, replaces all)")
	registryUpdateCmd.Flags().StringSliceVar(&regCustomValues, "custom-value", []string{}, "Custom key=value (repeatable, replaces all)")
	registryUpdateCmd.Flags().BoolVar(&regClearAdmins, "clear-admins", false, "Clear all admins")
	registryUpdateCmd.Flags().BoolVar(&regClearCustomVal, "clear-custom-values", false, "Clear all custom values")

	rootCmd.AddCommand(registryCmd)
}

func getAuthenticatedClient() *client.Client {
	serverURL, err := config.ResolveURL(flagURL)
	if err != nil {
		errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
	}

	token, err := auth.ResolveToken(flagToken)
	if err != nil {
		errors.ExitWithError(err, "failed to resolve authentication token")
	}

	// Send credentials to server if available; server determines if authentication is required
	var encodedToken string
	if token != "" {
		encodedToken = base64.StdEncoding.EncodeToString([]byte(token))
	}
	return client.NewClient(serverURL, encodedToken, flagTimeout, flagVerbose)
}

func runRegistryCreate(cmd *cobra.Command, args []string) {
	name := args[0]
	c := getAuthenticatedClient()

	// Validate and parse custom values
	customValues, err := validation.ParseCustomValues(regCustomValues)
	if err != nil {
		errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
	}

	// Build request
	reqBody := map[string]interface{}{
		"name": name,
	}
	if regDescription != "" {
		reqBody["description"] = regDescription
	}
	if len(regAdmins) > 0 {
		reqBody["admins"] = regAdmins
	}
	if len(customValues) > 0 {
		reqBody["custom_values"] = customValues
	}

	resp, err := c.Post("/api/v1/registry", reqBody)
	if err != nil {
		errors.ExitWithError(err, "failed to create registry")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to create registry: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]string{"name": name}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Created registry '%s'", name))
	}
}

func runRegistryList(cmd *cobra.Command, args []string) {
	c := getAuthenticatedClient()

	resp, err := c.Get("/api/v1/registry")
	if err != nil {
		errors.ExitWithError(err, "failed to list registries")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to list registries: %s", string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ExitWithError(err, "failed to read response")
	}

	var registries []map[string]interface{}
	if err := json.Unmarshal(body, &registries); err != nil {
		errors.ExitWithError(err, "failed to parse response")
	}

	if flagJSON {
		output.OutputJSON(registries, nil)
	} else {
		if len(registries) == 0 {
			fmt.Println("No registries found")
			return
		}

		table := output.NewTableWriter()
		table.WriteHeader("NAME", "DESCRIPTION", "PACKAGES")
		for _, reg := range registries {
			name := fmt.Sprintf("%v", reg["name"])
			description := fmt.Sprintf("%v", reg["description"])
			packages := "0"
			// Packages are returned as a map, not array
			if pkgs, ok := reg["packages"].(map[string]interface{}); ok {
				packages = strconv.Itoa(len(pkgs))
			}
			table.WriteRow(name, description, packages)
		}
		table.Flush()
	}
}

func runRegistryGet(cmd *cobra.Command, args []string) {
	name := args[0]
	c := getAuthenticatedClient()

	resp, err := c.Get("/api/v1/registry/" + name)
	if err != nil {
		errors.ExitWithError(err, "failed to get registry")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to get registry: %s", string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errors.ExitWithError(err, "failed to read response")
	}

	var registry map[string]interface{}
	if err := json.Unmarshal(body, &registry); err != nil {
		errors.ExitWithError(err, "failed to parse response")
	}

	if flagJSON {
		output.OutputJSON(registry, nil)
	} else {
		fmt.Printf("Name: %v\n", registry["name"])
		fmt.Printf("Description: %v\n", registry["description"])
		if admins, ok := registry["admins"].([]interface{}); ok && len(admins) > 0 {
			fmt.Print("Admins:")
			for _, admin := range admins {
				fmt.Printf("\n  - %v", admin)
			}
			fmt.Println()
		}
		if customVals, ok := registry["custom_values"].(map[string]interface{}); ok && len(customVals) > 0 {
			fmt.Println("Custom Values:")
			for k, v := range customVals {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}
}

func runRegistryUpdate(cmd *cobra.Command, args []string) {
	name := args[0]
	c := getAuthenticatedClient()

	// Validate flag conflicts
	if regClearAdmins && len(regAdmins) > 0 {
		errors.ExitWithCode(errors.ExitInvalidArguments, "cannot use --clear-admins with --admin. Use one or the other")
	}
	if regClearCustomVal && len(regCustomValues) > 0 {
		errors.ExitWithCode(errors.ExitInvalidArguments, "cannot use --clear-custom-values with --custom-value. Use one or the other")
	}

	// Validate and parse custom values
	var customValues map[string]string
	if len(regCustomValues) > 0 {
		var err error
		customValues, err = validation.ParseCustomValues(regCustomValues)
		if err != nil {
			errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
		}
	}

	// Build partial update request
	reqBody := make(map[string]interface{})
	if regDescription != "" {
		reqBody["description"] = regDescription
	}
	if regClearAdmins {
		reqBody["admins"] = []string{}
	} else if len(regAdmins) > 0 {
		reqBody["admins"] = regAdmins
	}
	if regClearCustomVal {
		reqBody["custom_values"] = map[string]string{}
	} else if len(customValues) > 0 {
		reqBody["custom_values"] = customValues
	}

	resp, err := c.Put("/api/v1/registry/"+name, reqBody)
	if err != nil {
		errors.ExitWithError(err, "failed to update registry")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to update registry: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]string{"name": name}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Updated registry '%s'", name))
	}
}

func runRegistryDelete(cmd *cobra.Command, args []string) {
	name := args[0]
	c := getAuthenticatedClient()

	// Prompt for confirmation unless --yes flag is set
	if !flagYes {
		if !prompts.ConfirmDeletion("registry", name, "all its packages and versions") {
			fmt.Println("Deletion cancelled")
			return
		}
	}

	resp, err := c.Delete("/api/v1/registry/" + name)
	if err != nil {
		errors.ExitWithError(err, "failed to delete registry")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("failed to delete registry: %s", string(body)))
	}

	if flagJSON {
		output.OutputJSON(map[string]bool{"deleted": true}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Deleted registry '%s'", name))
	}
}
