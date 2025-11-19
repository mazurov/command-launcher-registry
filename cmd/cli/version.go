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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage package versions",
}

var versionPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a new version",
	Run:   publishVersion,
}

var versionListCmd = &cobra.Command{
	Use:   "list [registry] [package]",
	Short: "List versions of a package",
	Args:  cobra.ExactArgs(2),
	Run:   listVersions,
}

var versionDeleteCmd = &cobra.Command{
	Use:   "delete [registry] [package] [version]",
	Short: "Delete a version",
	Args:  cobra.ExactArgs(3),
	Run:   deleteVersion,
}

var (
	verRegistry       string
	verPackage        string
	verVersion        string
	verURL            string
	verChecksum       string
	verStartPartition uint8
	verEndPartition   uint8
)

func init() {
	versionPublishCmd.Flags().StringVar(&verRegistry, "registry", "", "Registry name (required)")
	versionPublishCmd.Flags().StringVar(&verPackage, "package", "", "Package name (required)")
	versionPublishCmd.Flags().StringVar(&verVersion, "version", "", "Version string (required)")
	versionPublishCmd.Flags().StringVar(&verURL, "url", "", "Package URL (required)")
	versionPublishCmd.Flags().StringVar(&verChecksum, "checksum", "", "Package checksum")
	versionPublishCmd.Flags().Uint8Var(&verStartPartition, "start-partition", 0, "Start partition")
	versionPublishCmd.Flags().Uint8Var(&verEndPartition, "end-partition", 9, "End partition")
	_ = versionPublishCmd.MarkFlagRequired("registry")
	_ = versionPublishCmd.MarkFlagRequired("package")
	_ = versionPublishCmd.MarkFlagRequired("version")
	_ = versionPublishCmd.MarkFlagRequired("url")

	versionCmd.AddCommand(versionPublishCmd)
	versionCmd.AddCommand(versionListCmd)
	versionCmd.AddCommand(versionDeleteCmd)
}

func publishVersion(cmd *cobra.Command, args []string) {
	req := types.PublishVersionRequest{
		Version:        verVersion,
		URL:            verURL,
		Checksum:       verChecksum,
		StartPartition: verStartPartition,
		EndPartition:   verEndPartition,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		return
	}

	path := fmt.Sprintf("/remote/registries/%s/packages/%s/versions", verRegistry, verPackage)
	resp, err := doRequest("POST", path, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error publishing version: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("Version '%s' published successfully for package '%s/%s'\n", verVersion, verRegistry, verPackage)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to publish version: %s\n", string(body))
	}
}

func listVersions(cmd *cobra.Command, args []string) {
	registry := args[0]
	pkg := args[1]
	path := fmt.Sprintf("/remote/registries/%s/packages/%s/versions", registry, pkg)
	resp, err := doRequest("GET", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing versions: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func deleteVersion(cmd *cobra.Command, args []string) {
	registry := args[0]
	pkg := args[1]
	version := args[2]
	path := fmt.Sprintf("/remote/registries/%s/packages/%s/versions/%s", registry, pkg, version)
	resp, err := doRequest("DELETE", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting version: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Version '%s' deleted successfully from package '%s/%s'\n", version, registry, pkg)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Failed to delete version: %s\n", string(body))
	}
}
