/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package datastore

import (
	"os"
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
}
