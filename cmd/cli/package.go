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

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage packages",
}

var packageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new package",
	Run:   createPackage,
}

var packageListCmd = &cobra.Command{
	Use:   "list [registry]",
	Short: "List packages in a registry",
	Args:  cobra.ExactArgs(1),
	Run:   listPackages,
}

var packageDeleteCmd = &cobra.Command{
	Use:   "delete [registry] [package]",
	Short: "Delete a package",
	Args:  cobra.ExactArgs(2),
	Run:   deletePackage,
}

var (
	pkgRegistry    string
	pkgName        string
	pkgDescription string
	pkgAdmin       []string
)

func init() {
	packageCreateCmd.Flags().StringVar(&pkgRegistry, "registry", "", "Registry name (required)")
	packageCreateCmd.Flags().StringVar(&pkgName, "name", "", "Package name (required)")
	packageCreateCmd.Flags().StringVar(&pkgDescription, "description", "", "Package description")
	packageCreateCmd.Flags().StringSliceVar(&pkgAdmin, "admin", []string{}, "Package administrators")
	_ = packageCreateCmd.MarkFlagRequired("registry")
	_ = packageCreateCmd.MarkFlagRequired("name")

	packageCmd.AddCommand(packageCreateCmd)
	packageCmd.AddCommand(packageListCmd)
	packageCmd.AddCommand(packageDeleteCmd)
}

func createPackage(cmd *cobra.Command, args []string) {
	req := types.CreatePackageRequest{
		Name:        pkgName,
		Description: pkgDescription,
		Admin:       pkgAdmin,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		return
	}

	path := fmt.Sprintf("/remote/registries/%s/packages", pkgRegistry)
	resp, err := doRequest("POST", path, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating package: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("Package '%s' created successfully in registry '%s'\n", pkgName, pkgRegistry)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to create package: %s\n", string(body))
	}
}

func listPackages(cmd *cobra.Command, args []string) {
	registry := args[0]
	path := fmt.Sprintf("/remote/registries/%s/packages", registry)
	resp, err := doRequest("GET", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing packages: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func deletePackage(cmd *cobra.Command, args []string) {
	registry := args[0]
	pkgName := args[1]
	path := fmt.Sprintf("/remote/registries/%s/packages/%s", registry, pkgName)
	resp, err := doRequest("DELETE", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting package: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Package '%s' deleted successfully from registry '%s'\n", pkgName, registry)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to delete package: %s\n", string(body))
	}
}
