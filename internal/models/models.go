package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringSlice is a custom type for storing string arrays in the database
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// StringMap is a custom type for storing key-value pairs in the database
type StringMap map[string]string

func (m StringMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(m)
}

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = map[string]string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Registry represents a package registry
type Registry struct {
	ID           uint           `gorm:"primaryKey" json:"-"`
	Name         string         `gorm:"uniqueIndex;not null" json:"name"`
	Description  string         `json:"description"`
	Admin        StringSlice    `gorm:"type:jsonb" json:"admin"`
	CustomValues StringMap      `gorm:"type:jsonb" json:"customValues"`
	Packages     []Package      `gorm:"foreignKey:RegistryID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Package represents a package within a registry
type Package struct {
	ID           uint           `gorm:"primaryKey" json:"-"`
	RegistryID   uint           `gorm:"index;not null" json:"-"`
	Registry     *Registry      `gorm:"foreignKey:RegistryID" json:"-"`
	Name         string         `gorm:"index;not null" json:"name"`
	Description  string         `json:"description"`
	Admin        StringSlice    `gorm:"type:jsonb" json:"admin"`
	CustomValues StringMap      `gorm:"type:jsonb" json:"customValues"`
	Versions     []Version      `gorm:"foreignKey:PackageID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Version represents a specific version of a package
type Version struct {
	ID             uint           `gorm:"primaryKey" json:"-"`
	PackageID      uint           `gorm:"index;not null" json:"-"`
	Package        *Package       `gorm:"foreignKey:PackageID" json:"-"`
	Version        string         `gorm:"index;not null" json:"version"`
	URL            string         `gorm:"not null" json:"url"`
	Checksum       string         `json:"checksum"`
	StartPartition uint8          `gorm:"default:0" json:"startPartition"`
	EndPartition   uint8          `gorm:"default:9" json:"endPartition"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides
func (Registry) TableName() string {
	return "registries"
}

func (Package) TableName() string {
	return "packages"
}

func (Version) TableName() string {
	return "versions"
}

// Ensure unique constraint on registry-package combination
type PackageIndex struct {
	RegistryID uint   `gorm:"uniqueIndex:idx_registry_package"`
	Name       string `gorm:"uniqueIndex:idx_registry_package"`
}

// Ensure unique constraint on package-version combination
type VersionIndex struct {
	PackageID uint   `gorm:"uniqueIndex:idx_package_version"`
	Version   string `gorm:"uniqueIndex:idx_package_version"`
}
