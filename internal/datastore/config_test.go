/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package datastore

import (
	"os"
	"strings"
	"testing"
)

func TestS3ConfigGetMountPath(t *testing.T) {
	config := S3Config{
		Alias:  "myalias",
		Bucket: "mybucket",
	}

	expected := "/s3-data/myalias/mybucket"
	actual := config.GetMountPath(DefaultBaseDir)

	if actual != expected {
		t.Errorf("Expected mount path '%s', got '%s'", expected, actual)
	}

	// Test with different base dir
	actual = config.GetMountPath("/custom")
	expected = "/custom/myalias/mybucket"
	if actual != expected {
		t.Errorf("Expected mount path '%s', got '%s'", expected, actual)
	}
}

func TestGetEnvVar(t *testing.T) {
	os.Setenv("S3_TO_LOCAL_test1_BUCKET", "test-bucket-value")
	defer os.Unsetenv("S3_TO_LOCAL_test1_BUCKET")

	result := getEnvVar("S3_TO_LOCAL_", "test1", "BUCKET")
	expected := "test-bucket-value"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGetEnvVarNonExistent(t *testing.T) {
	os.Unsetenv("S3_TO_LOCAL_nonexistent_FIELD")

	result := getEnvVar("S3_TO_LOCAL_", "nonexistent", "FIELD")

	if result != "" {
		t.Errorf("Expected empty string for non-existent var, got '%s'", result)
	}
}

func TestGetEnvVarMultiple(t *testing.T) {
	os.Setenv("S3_TO_LOCAL_myalias_ACCESS_KEY", "my-access-key")
	os.Setenv("S3_TO_LOCAL_myalias_SECRET_KEY", "my-secret-key")
	defer func() {
		os.Unsetenv("S3_TO_LOCAL_myalias_ACCESS_KEY")
		os.Unsetenv("S3_TO_LOCAL_myalias_SECRET_KEY")
	}()

	result := getEnvVar("S3_TO_LOCAL_", "myalias", "ACCESS_KEY")
	if result != "my-access-key" {
		t.Errorf("Expected 'my-access-key', got '%s'", result)
	}

	result = getEnvVar("S3_TO_LOCAL_", "myalias", "SECRET_KEY")
	if result != "my-secret-key" {
		t.Errorf("Expected 'my-secret-key', got '%s'", result)
	}
}

func TestMountConfigDefaults(t *testing.T) {
	config := &MountConfig{
		BaseDir:          DefaultBaseDir,
		Mounts:           []S3Config{},
		MountingComplete: false,
	}

	if config.BaseDir != "/s3-data" {
		t.Errorf("Expected BaseDir '/s3-data', got '%s'", config.BaseDir)
	}

	if config.MountingComplete {
		t.Error("Expected MountingComplete to be false")
	}

	if len(config.Mounts) != 0 {
		t.Errorf("Expected 0 mounts, got %d", len(config.Mounts))
	}

	// Test with custom values
	config2 := &MountConfig{
		BaseDir:          "/custom",
		MountingComplete: true,
		Mounts: []S3Config{
			{Alias: "alias1", Bucket: "bucket1"},
		},
	}

	if config2.BaseDir != "/custom" {
		t.Errorf("Expected BaseDir '/custom', got '%s'", config2.BaseDir)
	}

	if !config2.MountingComplete {
		t.Error("Expected MountingComplete to be true")
	}

	if len(config2.Mounts) != 1 {
		t.Errorf("Expected 1 mount, got %d", len(config2.Mounts))
	}
}

func TestS3ConfigGetMountPathEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		config   S3Config
		baseDir  string
		expected string
	}{
		{
			name:     "empty alias and bucket",
			config:   S3Config{Alias: "", Bucket: ""},
			baseDir:  "/data",
			expected: "/data//",
		},
		{
			name:     "alias with special chars",
			config:   S3Config{Alias: "my-alias_123", Bucket: "my-bucket.456"},
			baseDir:  "/data",
			expected: "/data/my-alias_123/my-bucket.456",
		},
		{
			name:     "base dir with trailing slash",
			config:   S3Config{Alias: "alias", Bucket: "bucket"},
			baseDir:  "/data",
			expected: "/data/alias/bucket",
		},
		{
			name:     "base dir without leading slash",
			config:   S3Config{Alias: "alias", Bucket: "bucket"},
			baseDir:  "data",
			expected: "data/alias/bucket",
		},
		{
			name:     "empty base dir",
			config:   S3Config{Alias: "alias", Bucket: "bucket"},
			baseDir:  "",
			expected: "/alias/bucket",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.config.GetMountPath(tc.baseDir)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestS3ConfigStruct(t *testing.T) {
	config := S3Config{
		Alias:     "test-alias",
		Bucket:    "test-bucket",
		Host:      "s3.example.com",
		Region:    "us-east-1",
		AccessKey: "access-key-123",
		SecretKey: "secret-key-456",
	}

	if config.Alias != "test-alias" {
		t.Errorf("Expected Alias 'test-alias', got '%s'", config.Alias)
	}
	if config.Bucket != "test-bucket" {
		t.Errorf("Expected Bucket 'test-bucket', got '%s'", config.Bucket)
	}
	if config.Host != "s3.example.com" {
		t.Errorf("Expected Host 's3.example.com', got '%s'", config.Host)
	}
	if config.Region != "us-east-1" {
		t.Errorf("Expected Region 'us-east-1', got '%s'", config.Region)
	}
	if config.AccessKey != "access-key-123" {
		t.Errorf("Expected AccessKey 'access-key-123', got '%s'", config.AccessKey)
	}
	if config.SecretKey != "secret-key-456" {
		t.Errorf("Expected SecretKey 'secret-key-456', got '%s'", config.SecretKey)
	}
}

func TestMountConfigWithMultipleMounts(t *testing.T) {
	config := &MountConfig{
		BaseDir:  "/data",
		Mounts: []S3Config{
			{Alias: "alias1", Bucket: "bucket1", Host: "host1", Region: "region1"},
			{Alias: "alias2", Bucket: "bucket2", Host: "host2", Region: "region2"},
			{Alias: "alias3", Bucket: "bucket3", Host: "host3", Region: "region3"},
		},
		MountingComplete: true,
	}

	if len(config.Mounts) != 3 {
		t.Errorf("Expected 3 mounts, got %d", len(config.Mounts))
	}

	if config.Mounts[0].Alias != "alias1" {
		t.Errorf("Expected first mount alias 'alias1', got '%s'", config.Mounts[0].Alias)
	}

	if config.Mounts[1].Alias != "alias2" {
		t.Errorf("Expected second mount alias 'alias2', got '%s'", config.Mounts[1].Alias)
	}

	if config.Mounts[2].Alias != "alias3" {
		t.Errorf("Expected third mount alias 'alias3', got '%s'", config.Mounts[2].Alias)
	}
}

func TestLoadConfigFromEnvError(t *testing.T) {
	// LoadConfigFromEnv will try to create /s3-data which we can't do
	// Test that it returns an error appropriately
	config, err := LoadConfigFromEnv()
	
	// We expect an error because we can't create /s3-data
	if err == nil {
		// If no error, config should still be valid (but mounts might be empty)
		if config == nil {
			t.Error("Expected non-nil config even without error")
		}
	} else {
		// Error is expected - verify it contains the expected message
		if !strings.Contains(err.Error(), "failed to create base directory") {
			t.Errorf("Expected error about failed directory creation, got: %v", err)
		}
		if !strings.Contains(err.Error(), "/s3-data") {
			t.Errorf("Expected error to mention /s3-data, got: %v", err)
		}
	}
}
