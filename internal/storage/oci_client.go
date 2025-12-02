package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// OCI timeout constants per FR-016
const (
	OCIPushTimeout = 60 * time.Second  // Increased from 5s - ghcr.io can be slow
	OCIPullTimeout = 30 * time.Second
)

// OCI media types for registry data artifact
const (
	OCIConfigMediaType = "application/vnd.oci.image.config.v1+json"
	OCILayerMediaType  = "application/json"
	OCIManifestTitle   = "registry.json"
)

// OCIClient wraps oras-go for OCI registry operations
type OCIClient struct {
	repository *remote.Repository
	reference  string // Full reference "registry/repo:latest"
	logger     *slog.Logger
}

// NewOCIClient creates a new OCI client for the given reference and token.
// The reference should be in format "registry/repo:tag" (e.g., "ghcr.io/org/repo:latest").
// The token is used as a bearer token for authentication.
func NewOCIClient(reference string, token string, logger *slog.Logger) (*OCIClient, error) {
	start := time.Now()

	repo, err := remote.NewRepository(reference)
	if err != nil {
		logger.Error("Failed to create OCI repository",
			"reference", reference,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeOCIError(OCIOpConnect, fmt.Errorf("invalid OCI reference %q: %w", reference, err))
	}

	// Configure authentication
	// Use token as password with empty username - this works for:
	// - ghcr.io: GitHub PAT as password (username can be anything non-empty)
	// - docker.io: access token as password
	// - ACR/ECR: tokens work as password with special usernames
	if token != "" {
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Credential: func(ctx context.Context, reg string) (auth.Credential, error) {
				return auth.Credential{
					Username: "token",
					Password: token,
				}, nil
			},
		}
	}

	logger.Info("OCI client created",
		"reference", reference,
		"has_token", token != "",
		"duration_ms", time.Since(start).Milliseconds())

	return &OCIClient{
		repository: repo,
		reference:  reference,
		logger:     logger,
	}, nil
}

// Pull retrieves the registry data from the OCI repository.
// Uses 30s timeout per FR-016. Returns the JSON data or an error.
func (c *OCIClient) Pull(ctx context.Context) ([]byte, error) {
	start := time.Now()
	c.logger.Debug("Starting OCI pull", "reference", c.reference)

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, OCIPullTimeout)
	defer cancel()

	// Create in-memory store for pulled content
	store := memory.New()

	// Pull the artifact
	desc, err := oras.Copy(ctx, c.repository, c.repository.Reference.Reference, store, "", oras.DefaultCopyOptions)
	if err != nil {
		c.logger.Error("OCI pull failed",
			"reference", c.reference,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeOCIError(OCIOpPull, err)
	}

	// Fetch the manifest
	manifestReader, err := store.Fetch(ctx, desc)
	if err != nil {
		c.logger.Error("Failed to fetch manifest",
			"reference", c.reference,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeOCIError(OCIOpPull, fmt.Errorf("failed to fetch manifest: %w", err))
	}
	defer manifestReader.Close()

	var manifestBuf bytes.Buffer
	if _, err := manifestBuf.ReadFrom(manifestReader); err != nil {
		return nil, CategorizeOCIError(OCIOpPull, fmt.Errorf("failed to read manifest: %w", err))
	}

	// Parse manifest to find the data layer
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBuf.Bytes(), &manifest); err != nil {
		c.logger.Error("Failed to parse manifest",
			"reference", c.reference,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeOCIError(OCIOpPull, fmt.Errorf("failed to parse manifest: %w", err))
	}

	// Find the registry.json layer
	if len(manifest.Layers) == 0 {
		return nil, CategorizeOCIError(OCIOpPull, fmt.Errorf("artifact has no layers"))
	}

	// Get the first layer (registry.json data)
	layerDesc := manifest.Layers[0]
	layerReader, err := store.Fetch(ctx, layerDesc)
	if err != nil {
		c.logger.Error("Failed to fetch layer",
			"reference", c.reference,
			"digest", layerDesc.Digest,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return nil, CategorizeOCIError(OCIOpPull, fmt.Errorf("failed to fetch data layer: %w", err))
	}
	defer layerReader.Close()

	var dataBuf bytes.Buffer
	if _, err := dataBuf.ReadFrom(layerReader); err != nil {
		return nil, CategorizeOCIError(OCIOpPull, fmt.Errorf("failed to read data layer: %w", err))
	}

	data := dataBuf.Bytes()
	c.logger.Info("OCI pull completed",
		"reference", c.reference,
		"size_bytes", len(data),
		"duration_ms", time.Since(start).Milliseconds())

	return data, nil
}

// Push uploads the registry data to the OCI repository.
// Uses 60s timeout. Always uses the "latest" tag.
func (c *OCIClient) Push(ctx context.Context, data []byte) error {
	start := time.Now()
	c.logger.Info("Starting OCI push",
		"reference", c.reference,
		"size_bytes", len(data))

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, OCIPushTimeout)
	defer cancel()

	// Create in-memory store for the artifact
	store := memory.New()

	// Create the empty config blob
	configData := []byte("{}")
	configDesc := ocispec.Descriptor{
		MediaType: OCIConfigMediaType,
		Digest:    digest.FromBytes(configData),
		Size:      int64(len(configData)),
	}
	if err := store.Push(ctx, configDesc, bytes.NewReader(configData)); err != nil {
		return CategorizeOCIError(OCIOpPush, fmt.Errorf("failed to push config: %w", err))
	}

	// Create the data layer with annotations
	layerDesc := ocispec.Descriptor{
		MediaType: OCILayerMediaType,
		Digest:    digest.FromBytes(data),
		Size:      int64(len(data)),
		Annotations: map[string]string{
			ocispec.AnnotationTitle: OCIManifestTitle,
		},
	}
	if err := store.Push(ctx, layerDesc, bytes.NewReader(data)); err != nil {
		return CategorizeOCIError(OCIOpPush, fmt.Errorf("failed to push layer: %w", err))
	}

	// Create the manifest
	manifest := ocispec.Manifest{
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    configDesc,
		Layers:    []ocispec.Descriptor{layerDesc},
		Annotations: map[string]string{
			ocispec.AnnotationCreated:   time.Now().UTC().Format(time.RFC3339),
			"com.cola-registry.version": "1.0.0",
		},
	}
	manifest.SchemaVersion = 2

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return CategorizeOCIError(OCIOpPush, fmt.Errorf("failed to marshal manifest: %w", err))
	}

	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestJSON),
		Size:      int64(len(manifestJSON)),
	}
	if err := store.Push(ctx, manifestDesc, bytes.NewReader(manifestJSON)); err != nil {
		return CategorizeOCIError(OCIOpPush, fmt.Errorf("failed to push manifest: %w", err))
	}

	// Tag the manifest
	if err := store.Tag(ctx, manifestDesc, c.repository.Reference.Reference); err != nil {
		return CategorizeOCIError(OCIOpPush, fmt.Errorf("failed to tag manifest: %w", err))
	}

	// Copy to remote repository
	_, err = oras.Copy(ctx, store, c.repository.Reference.Reference, c.repository, "", oras.DefaultCopyOptions)
	if err != nil {
		c.logger.Error("OCI push failed",
			"reference", c.reference,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return CategorizeOCIError(OCIOpPush, err)
	}

	c.logger.Info("OCI push completed",
		"reference", c.reference,
		"size_bytes", len(data),
		"duration_ms", time.Since(start).Milliseconds())

	return nil
}

// Exists checks if the artifact exists in the OCI repository.
func (c *OCIClient) Exists(ctx context.Context) (bool, error) {
	start := time.Now()
	c.logger.Debug("Checking OCI artifact existence", "reference", c.reference)

	// Apply pull timeout for existence check
	ctx, cancel := context.WithTimeout(ctx, OCIPullTimeout)
	defer cancel()

	_, err := c.repository.Resolve(ctx, c.repository.Reference.Reference)
	if err != nil {
		// Check if it's a "not found" error
		errStr := err.Error()
		// oras-go returns various "not found" formats:
		// - "ghcr.io/user/repo:tag: not found"
		// - HTTP 404 status
		// - "NAME_UNKNOWN" or "MANIFEST_UNKNOWN" errors
		if containsHTTPStatus(errStr, 404) || containsHTTPStatus(errStr, 400) ||
			strings.HasSuffix(errStr, ": not found") ||
			strings.Contains(errStr, "NOT_FOUND") ||
			strings.Contains(errStr, "NAME_UNKNOWN") ||
			strings.Contains(errStr, "MANIFEST_UNKNOWN") {
			c.logger.Info("OCI artifact does not exist",
				"reference", c.reference,
				"duration_ms", time.Since(start).Milliseconds())
			return false, nil
		}
		c.logger.Error("OCI existence check failed",
			"reference", c.reference,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds())
		return false, CategorizeOCIError(OCIOpConnect, err)
	}

	c.logger.Info("OCI artifact exists",
		"reference", c.reference,
		"duration_ms", time.Since(start).Milliseconds())
	return true, nil
}
