package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/criteo/command-launcher-registry/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBaseStorage() *BaseStorage {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewBaseStorage(logger)
}

func TestBaseStorage_NewBaseStorage(t *testing.T) {
	bs := newTestBaseStorage()
	assert.NotNil(t, bs)
	assert.NotNil(t, bs.data)
	assert.NotNil(t, bs.data.Registries)
}

func TestBaseStorage_SetGetData(t *testing.T) {
	bs := newTestBaseStorage()

	// Create test data
	testData := &models.Storage{
		Registries: map[string]*models.Registry{
			"test-reg": {
				Name:     "test-reg",
				Packages: make(map[string]*models.Package),
			},
		},
	}

	// Set data
	bs.SetData(testData)

	// Get data
	data := bs.GetData()
	assert.Equal(t, 1, len(data.Registries))
	assert.Equal(t, "test-reg", data.Registries["test-reg"].Name)
}

func TestBaseStorage_MarshalUnmarshalData(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Create a registry
	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	// Marshal
	jsonData, err := bs.MarshalData()
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "test-reg")

	// Create new storage and unmarshal
	bs2 := newTestBaseStorage()
	err = bs2.UnmarshalData(jsonData)
	require.NoError(t, err)

	// Verify data
	data := bs2.GetData()
	assert.Equal(t, 1, len(data.Registries))
	assert.Equal(t, "test-reg", data.Registries["test-reg"].Name)
}

func TestBaseStorage_CreateRegistry(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	assert.NoError(t, err)

	// Verify created
	retrieved, err := bs.GetRegistry(ctx, "test-reg")
	require.NoError(t, err)
	assert.Equal(t, "test-reg", retrieved.Name)
	assert.Equal(t, "Test Registry", retrieved.Description)
}

func TestBaseStorage_CreateRegistry_AlreadyExists(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	// Try to create again
	err = bs.CreateRegistry(ctx, reg, nil)
	assert.ErrorIs(t, err, ErrAlreadyExists)
}

func TestBaseStorage_CreateRegistry_WithPersistCallback(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	persistCalled := false
	persistFunc := func() error {
		persistCalled = true
		return nil
	}

	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, persistFunc)
	assert.NoError(t, err)
	assert.True(t, persistCalled)
}

func TestBaseStorage_CreateRegistry_PersistFailure_Rollback(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	persistFunc := func() error {
		return assert.AnError
	}

	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, persistFunc)
	assert.ErrorIs(t, err, ErrStorageUnavailable)

	// Verify rollback - registry should not exist
	_, err = bs.GetRegistry(ctx, "test-reg")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestBaseStorage_GetRegistry_NotFound(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	_, err := bs.GetRegistry(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestBaseStorage_UpdateRegistry(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Create registry
	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	// Update registry
	updatedReg := models.NewRegistry("test-reg", "Updated Description", nil, nil)
	err = bs.UpdateRegistry(ctx, updatedReg, nil)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := bs.GetRegistry(ctx, "test-reg")
	require.NoError(t, err)
	assert.Equal(t, "Updated Description", retrieved.Description)
}

func TestBaseStorage_DeleteRegistry(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Create registry
	reg := models.NewRegistry("test-reg", "Test Registry", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	// Delete registry
	err = bs.DeleteRegistry(ctx, "test-reg", nil)
	assert.NoError(t, err)

	// Verify deleted
	_, err = bs.GetRegistry(ctx, "test-reg")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestBaseStorage_ListRegistries(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Create registries
	for i := 0; i < 3; i++ {
		reg := models.NewRegistry(string(rune('a'+i)), "", nil, nil)
		err := bs.CreateRegistry(ctx, reg, nil)
		require.NoError(t, err)
	}

	// List registries
	registries, err := bs.ListRegistries(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(registries))
}

func TestBaseStorage_CreatePackage(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Create registry first
	reg := models.NewRegistry("test-reg", "", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	// Create package
	pkg := models.NewPackage("test-pkg", "Test Package", nil, nil)
	err = bs.CreatePackage(ctx, "test-reg", pkg, nil)
	assert.NoError(t, err)

	// Verify created
	retrieved, err := bs.GetPackage(ctx, "test-reg", "test-pkg")
	require.NoError(t, err)
	assert.Equal(t, "test-pkg", retrieved.Name)
}

func TestBaseStorage_CreatePackage_RegistryNotFound(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	pkg := models.NewPackage("test-pkg", "", nil, nil)
	err := bs.CreatePackage(ctx, "nonexistent", pkg, nil)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestBaseStorage_CreateVersion(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Setup registry and package
	reg := models.NewRegistry("test-reg", "", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	pkg := models.NewPackage("test-pkg", "", nil, nil)
	err = bs.CreatePackage(ctx, "test-reg", pkg, nil)
	require.NoError(t, err)

	// Create version
	ver := &models.Version{
		Name:           "test-pkg",
		Version:        "1.0.0",
		Checksum:       "abc123",
		URL:            "http://example.com/pkg.zip",
		StartPartition: 0,
		EndPartition:   9,
	}
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver, nil)
	assert.NoError(t, err)

	// Verify created
	retrieved, err := bs.GetVersion(ctx, "test-reg", "test-pkg", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", retrieved.Version)
}

func TestBaseStorage_CreateVersion_Immutability(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Setup
	reg := models.NewRegistry("test-reg", "", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	pkg := models.NewPackage("test-pkg", "", nil, nil)
	err = bs.CreatePackage(ctx, "test-reg", pkg, nil)
	require.NoError(t, err)

	ver := &models.Version{
		Name:           "test-pkg",
		Version:        "1.0.0",
		StartPartition: 0,
		EndPartition:   9,
	}
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver, nil)
	require.NoError(t, err)

	// Try to create same version again
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver, nil)
	assert.ErrorIs(t, err, ErrImmutabilityViolation)
}

func TestBaseStorage_CreateVersion_PartitionOverlap(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Setup
	reg := models.NewRegistry("test-reg", "", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	pkg := models.NewPackage("test-pkg", "", nil, nil)
	err = bs.CreatePackage(ctx, "test-reg", pkg, nil)
	require.NoError(t, err)

	// Create first version with partitions 0-4
	ver1 := &models.Version{
		Name:           "test-pkg",
		Version:        "1.0.0",
		StartPartition: 0,
		EndPartition:   4,
	}
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver1, nil)
	require.NoError(t, err)

	// Try to create overlapping version with partitions 3-7
	ver2 := &models.Version{
		Name:           "test-pkg",
		Version:        "2.0.0",
		StartPartition: 3,
		EndPartition:   7,
	}
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver2, nil)
	assert.ErrorIs(t, err, ErrPartitionOverlap)
}

func TestBaseStorage_DeleteVersion(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Setup
	reg := models.NewRegistry("test-reg", "", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	pkg := models.NewPackage("test-pkg", "", nil, nil)
	err = bs.CreatePackage(ctx, "test-reg", pkg, nil)
	require.NoError(t, err)

	ver := &models.Version{
		Name:           "test-pkg",
		Version:        "1.0.0",
		StartPartition: 0,
		EndPartition:   9,
	}
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver, nil)
	require.NoError(t, err)

	// Delete version
	err = bs.DeleteVersion(ctx, "test-reg", "test-pkg", "1.0.0", nil)
	assert.NoError(t, err)

	// Verify deleted
	_, err = bs.GetVersion(ctx, "test-reg", "test-pkg", "1.0.0")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestBaseStorage_GetRegistryIndex(t *testing.T) {
	bs := newTestBaseStorage()
	ctx := context.Background()

	// Setup
	reg := models.NewRegistry("test-reg", "", nil, nil)
	err := bs.CreateRegistry(ctx, reg, nil)
	require.NoError(t, err)

	pkg := models.NewPackage("test-pkg", "", nil, nil)
	err = bs.CreatePackage(ctx, "test-reg", pkg, nil)
	require.NoError(t, err)

	ver := &models.Version{
		Name:           "test-pkg",
		Version:        "1.0.0",
		Checksum:       "abc123",
		URL:            "http://example.com/pkg.zip",
		StartPartition: 0,
		EndPartition:   9,
	}
	err = bs.CreateVersion(ctx, "test-reg", "test-pkg", ver, nil)
	require.NoError(t, err)

	// Get index
	entries, err := bs.GetRegistryIndex(ctx, "test-reg")
	require.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "test-pkg", entries[0].Name)
	assert.Equal(t, "1.0.0", entries[0].Version)
}
