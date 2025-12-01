package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// tests/integration/server_config_test.go -> project root
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// TestServerStartsWithStorageURI verifies the server starts with --storage-uri flag
func TestServerStartsWithStorageURI(t *testing.T) {
	projectRoot := getProjectRoot()

	// Create temp directory for test data
	tmpDir, err := os.MkdirTemp("", "cola-registry-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storageFile := filepath.Join(tmpDir, "registry.json")
	storageURI := "file://" + storageFile

	// Find an available port
	port := 18080

	// Build the server binary
	binPath := filepath.Join(tmpDir, "cola-registry")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/cola-registry")
	buildCmd.Dir = projectRoot
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server: %s", string(output))

	// Start server with storage URI
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverCmd := exec.CommandContext(ctx, binPath, "server",
		"--storage-uri", storageURI,
		"--port", fmt.Sprintf("%d", port),
	)
	serverCmd.Dir = tmpDir

	err = serverCmd.Start()
	require.NoError(t, err, "Failed to start server")

	// Ensure server is cleaned up
	defer func() {
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()

	// Wait for server to be ready
	serverReady := waitForServer(fmt.Sprintf("http://localhost:%d/api/v1/health", port), 5*time.Second)
	assert.True(t, serverReady, "Server should be ready within timeout")

	// Verify storage file was created
	_, err = os.Stat(storageFile)
	assert.NoError(t, err, "Storage file should be created")
}

// TestServerStartsWithEnvVars verifies the server starts with environment variables only
func TestServerStartsWithEnvVars(t *testing.T) {
	projectRoot := getProjectRoot()

	// Create temp directory for test data
	tmpDir, err := os.MkdirTemp("", "cola-registry-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storageFile := filepath.Join(tmpDir, "registry.json")
	storageURI := "file://" + storageFile

	// Find an available port
	port := 18081

	// Build the server binary
	binPath := filepath.Join(tmpDir, "cola-registry")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/cola-registry")
	buildCmd.Dir = projectRoot
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server: %s", string(output))

	// Start server with environment variables only
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverCmd := exec.CommandContext(ctx, binPath, "server")
	serverCmd.Dir = tmpDir
	serverCmd.Env = append(os.Environ(),
		"COLA_REGISTRY_STORAGE_URI="+storageURI,
		fmt.Sprintf("COLA_REGISTRY_SERVER_PORT=%d", port),
	)

	err = serverCmd.Start()
	require.NoError(t, err, "Failed to start server")

	// Ensure server is cleaned up
	defer func() {
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()

	// Wait for server to be ready
	serverReady := waitForServer(fmt.Sprintf("http://localhost:%d/api/v1/health", port), 5*time.Second)
	assert.True(t, serverReady, "Server should be ready within timeout")

	// Verify storage file was created
	_, err = os.Stat(storageFile)
	assert.NoError(t, err, "Storage file should be created")
}

// TestServerStartsWithPathWithoutScheme verifies auto-prefix of file:// works
func TestServerStartsWithPathWithoutScheme(t *testing.T) {
	projectRoot := getProjectRoot()

	// Create temp directory for test data
	tmpDir, err := os.MkdirTemp("", "cola-registry-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Use path without scheme - should auto-prefix with file://
	storageFile := filepath.Join(tmpDir, "registry.json")

	// Find an available port
	port := 18082

	// Build the server binary
	binPath := filepath.Join(tmpDir, "cola-registry")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/cola-registry")
	buildCmd.Dir = projectRoot
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server: %s", string(output))

	// Start server with path without scheme
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverCmd := exec.CommandContext(ctx, binPath, "server",
		"--storage-uri", storageFile, // No file:// prefix
		"--port", fmt.Sprintf("%d", port),
	)
	serverCmd.Dir = tmpDir

	err = serverCmd.Start()
	require.NoError(t, err, "Failed to start server")

	// Ensure server is cleaned up
	defer func() {
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()

	// Wait for server to be ready
	serverReady := waitForServer(fmt.Sprintf("http://localhost:%d/api/v1/health", port), 5*time.Second)
	assert.True(t, serverReady, "Server should be ready within timeout")

	// Verify storage file was created
	_, err = os.Stat(storageFile)
	assert.NoError(t, err, "Storage file should be created")
}

// waitForServer waits for the server to be ready
func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 1 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
