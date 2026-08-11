/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package datastore

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// S3Config holds the configuration for a single S3 mount
type S3Config struct {
	// Alias is the user-defined identifier for this mount
	Alias string
	// Bucket is the S3 bucket name
	Bucket string
	// Host is the S3 endpoint host URL
	Host string
	// Region is the AWS region
	Region string
	// AccessKey is the S3 access key
	AccessKey string
	// SecretKey is the S3 secret key
	SecretKey string
}

// MountConfig holds the base configuration for mounting
type MountConfig struct {
	// BaseDir is the base directory where S3 buckets will be mounted
	BaseDir string
	// Mounts is the list of S3 configurations to mount
	Mounts []S3Config
	// MountingComplete indicates whether all mounts are ready
	MountingComplete bool
}

// DefaultBaseDir is the default base directory for S3 mounts
const DefaultBaseDir = "/s3-data"

// ActiveMountsFile is the file that tracks all active mount points
const ActiveMountsFile = "/etc/active_mounts.txt"

// EnvVarPattern is the regex pattern to find S3 alias environment variables
const EnvVarPattern = `^S3_TO_LOCAL_(.*)_ALIAS$`

// LoadConfigFromEnv loads S3 configurations from environment variables
// It scans for S3_TO_LOCAL_*_ALIAS variables and loads corresponding settings
func LoadConfigFromEnv() (*MountConfig, error) {
	config := &MountConfig{
		BaseDir:   DefaultBaseDir,
		Mounts:    []S3Config{},
		MountingComplete: false,
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", config.BaseDir, err)
	}

	// Find all S3_TO_LOCAL_*_ALIAS environment variables
	aliasPattern := regexp.MustCompile(EnvVarPattern)
	
	aliases := []string{}
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		if aliasPattern.MatchString(key) {
			matches := aliasPattern.FindStringSubmatch(key)
			if len(matches) > 1 {
				aliases = append(aliases, matches[1])
			}
		}
	}

	// For each alias, load the full configuration
	for _, alias := range aliases {
		s3Config := S3Config{
			Alias:     alias,
			Bucket:    getEnvVar("S3_TO_LOCAL_", alias, "BUCKET"),
			Host:      getEnvVar("S3_TO_LOCAL_", alias, "HOST"),
			Region:    getEnvVar("S3_TO_LOCAL_", alias, "REGION"),
			AccessKey: getEnvVar("S3_TO_LOCAL_", alias, "ACCESS_KEY"),
			SecretKey: getEnvVar("S3_TO_LOCAL_", alias, "SECRET_KEY"),
		}
		
		// Skip if bucket is empty (required field)
		if s3Config.Bucket == "" {
			continue
		}
		
		config.Mounts = append(config.Mounts, s3Config)
	}

	return config, nil
}

// getEnvVar retrieves an environment variable for a specific alias and field
func getEnvVar(prefix, alias, field string) string {
	key := fmt.Sprintf("%s%s_%s", prefix, alias, field)
	return os.Getenv(key)
}

// GetMountPath returns the full mount path for a given S3Config
func (c *S3Config) GetMountPath(baseDir string) string {
	return fmt.Sprintf("%s/%s/%s", baseDir, c.Alias, c.Bucket)
}
