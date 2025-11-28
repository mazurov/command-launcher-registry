package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mazurov/command-launcher-registry/pkg/types"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage registries",
}

var registryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new registry",
	Run:   createRegistry,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registries",
	Run:   listRegistries,
}

var registryGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get registry details",
	Args:  cobra.ExactArgs(1),
	Run:   getRegistry,
}

var registryDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a registry",
	Args:  cobra.ExactArgs(1),
	Run:   deleteRegistry,
}

var (
	regName        string
	regDescription string
	regAdmin       []string
)

func init() {
	registryCreateCmd.Flags().StringVar(&regName, "name", "", "Registry name (required)")
	registryCreateCmd.Flags().StringVar(&regDescription, "description", "", "Registry description")
	registryCreateCmd.Flags().StringSliceVar(&regAdmin, "admin", []string{}, "Registry administrators")
	_ = registryCreateCmd.MarkFlagRequired("name")

	registryCmd.AddCommand(registryCreateCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryGetCmd)
	registryCmd.AddCommand(registryDeleteCmd)
}

func createRegistry(cmd *cobra.Command, args []string) {
	req := types.CreateRegistryRequest{
		Name:        regName,
		Description: regDescription,
		Admin:       regAdmin,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		return
	}

	resp, err := doRequest("POST", "/remote/registries", bytes.NewReader(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating registry: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("Registry '%s' created successfully\n", regName)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to create registry: %s\n", string(body))
	}
}

func listRegistries(cmd *cobra.Command, args []string) {
	resp, err := doRequest("GET", "/remote/registries", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing registries: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func getRegistry(cmd *cobra.Command, args []string) {
	name := args[0]
	resp, err := doRequest("GET", fmt.Sprintf("/remote/registries/%s", name), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting registry: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func deleteRegistry(cmd *cobra.Command, args []string) {
	name := args[0]
	resp, err := doRequest("DELETE", fmt.Sprintf("/remote/registries/%s", name), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting registry: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Registry '%s' deleted successfully\n", name)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to delete registry: %s\n", string(body))
	}
}

func doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := apiURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Token priority cascade:
	// 1. --token flag (jwtToken variable)
	// 2. Environment variable COLA_REGISTRY_TOKEN
	// 3. Stored token from interactive login (~/.config/cola-registry/config.yaml)

	token := jwtToken

	// If no token from flag, check environment variable
	if token == "" {
		token = os.Getenv("COLA_REGISTRY_TOKEN")
	}

	// If still no token, try to load from storage
	if token == "" {
		storage, err := loadToken()
		if err == nil && storage != nil {
			token = storage.Token
		}
	}

	// Set Authorization header if token is available
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return http.DefaultClient.Do(req)
}
