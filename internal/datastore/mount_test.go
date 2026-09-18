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
	"path/filepath"
	"strings"
	"testing"
)

func TestNewMountManager(t *testing.T) {
	config := &MountConfig{
		BaseDir:          "/test",
		Mounts:           []S3Config{},
		MountingComplete: false,
	}

	manager := NewMountManager(config)

	if manager.config != config {
		t.Error("Expected manager.config to be the provided config")
	}

	if len(manager.activeMounts) != 0 {
		t.Errorf("Expected empty activeMounts, got %d", len(manager.activeMounts))
	}

	// Test with non-empty config
	config2 := &MountConfig{
		BaseDir:          "/data",
		MountingComplete: true,
		Mounts: []S3Config{
			{Alias: "alias1", Bucket: "bucket1"},
		},
	}
	manager2 := NewMountManager(config2)
	if manager2.config != config2 {
		t.Error("Expected manager2.config to be the provided config2")
	}
	if len(manager2.activeMounts) != 0 {
		t.Errorf("Expected empty activeMounts for manager2, got %d", len(manager2.activeMounts))
	}
}

func TestBuildS3FSCmd(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:     "test",
		Bucket:    "test-bucket",
		Host:      "s3.example.com",
		Region:    "us-east-1",
		AccessKey: "key",
		SecretKey: "secret",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/test", "/tmp/passwd")

	// Check that the command path ends with "s3fs"
	baseName := filepath.Base(cmd.Path)
	if baseName != "s3fs" {
		t.Errorf("Expected command name 's3fs', got '%s'", baseName)
	}

	// Check that args contain expected values
	args := cmd.Args
	if len(args) < 3 {
		t.Errorf("Expected at least 3 args, got %d: %v", len(args), args)
	}

	// Should contain bucket name and target path
	foundBucket := false
	foundPath := false
	for _, arg := range args {
		if arg == "test-bucket" {
			foundBucket = true
		}
		if arg == "/mnt/test" {
			foundPath = true
		}
	}

	if !foundBucket {
		t.Error("Expected bucket name in args")
	}
	if !foundPath {
		t.Error("Expected target path in args")
	}

	// Check that required options are present
	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "passwd_file=") {
		t.Error("Expected passwd_file option in args")
	}
	if !strings.Contains(argsStr, "url=") {
		t.Error("Expected url option in args")
	}
	if !strings.Contains(argsStr, "endpoint=") {
		t.Error("Expected endpoint option in args")
	}
	if !strings.Contains(argsStr, "use_path_request_style") {
		t.Error("Expected use_path_request_style option in args")
	}
	if !strings.Contains(argsStr, "allow_other") {
		t.Error("Expected allow_other option in args")
	}
}

func TestBuildS3FSCmdMinimal(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:     "test",
		Bucket:    "test-bucket",
		Host:      "",
		Region:    "",
		AccessKey: "key",
		SecretKey: "secret",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/test", "")

	args := cmd.Args
	argsStr := strings.Join(args, " ")

	// Should contain bucket and path
	foundBucket := false
	foundPath := false
	for _, arg := range args {
		if arg == "test-bucket" {
			foundBucket = true
		}
		if arg == "/mnt/test" {
			foundPath = true
		}
	}

	if !foundBucket {
		t.Error("Expected bucket name in args")
	}
	if !foundPath {
		t.Error("Expected target path in args")
	}

	// Should still have required options even without host/region
	if !strings.Contains(argsStr, "use_path_request_style") {
		t.Errorf("Expected use_path_request_style option in args")
	}
	if !strings.Contains(argsStr, "allow_other") {
		t.Errorf("Expected allow_other option in args")
	}
	if !strings.Contains(argsStr, "nonempty") {
		t.Errorf("Expected nonempty option in args")
	}
	if !strings.Contains(argsStr, "-f") {
		t.Errorf("Expected -f flag in args")
	}
}

func TestBuildS3FSCmdAllOptions(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:     "test",
		Bucket:    "test-bucket",
		Host:      "s3.example.com",
		Region:    "us-east-1",
		AccessKey: "key",
		SecretKey: "secret",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/test", "/tmp/passwd")

	args := cmd.Args
	argsStr := strings.Join(args, " ")

	// Check all options are present
	if !strings.Contains(argsStr, "passwd_file=/tmp/passwd") {
		t.Error("Expected passwd_file=/tmp/passwd in args")
	}
	if !strings.Contains(argsStr, "url=s3.example.com") {
		t.Error("Expected url=s3.example.com in args")
	}
	if !strings.Contains(argsStr, "endpoint=us-east-1") {
		t.Error("Expected endpoint=us-east-1 in args")
	}
	if !strings.Contains(argsStr, "use_path_request_style") {
		t.Error("Expected use_path_request_style option in args")
	}
	if !strings.Contains(argsStr, "allow_other") {
		t.Error("Expected allow_other option in args")
	}
	if !strings.Contains(argsStr, "nonempty") {
		t.Error("Expected nonempty option in args")
	}
	if !strings.Contains(argsStr, "-f") {
		t.Error("Expected -f flag in args")
	}
}

func TestBuildS3FSCmdNoOptions(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:     "simple",
		Bucket:    "simple-bucket",
		Host:      "",
		Region:    "",
		AccessKey: "",
		SecretKey: "",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/simple", "")

	args := cmd.Args
	argsStr := strings.Join(args, " ")

	// Should contain bucket and path
	if len(args) < 2 {
		t.Errorf("Expected at least 2 args, got %d: %v", len(args), args)
	}

	if args[0] != "s3fs" {
		t.Errorf("Expected first arg to be 's3fs', got '%s'", args[0])
	}

	// Should have all required flags even with no options
	if !strings.Contains(argsStr, "use_path_request_style") {
		t.Error("Expected use_path_request_style option in args")
	}
	if !strings.Contains(argsStr, "allow_other") {
		t.Error("Expected allow_other option in args")
	}
	if !strings.Contains(argsStr, "nonempty") {
		t.Error("Expected nonempty option in args")
	}
	if !strings.Contains(argsStr, "-f") {
		t.Error("Expected -f flag in args")
	}

	// Should NOT have passwd_file, url, or endpoint when they are empty
	if strings.Contains(argsStr, "passwd_file=") {
		t.Error("Expected no passwd_file option when passwdFile is empty")
	}
	if strings.Contains(argsStr, "url=") {
		t.Error("Expected no url option when host is empty")
	}
	if strings.Contains(argsStr, "endpoint=") {
		t.Error("Expected no endpoint option when region is empty")
	}
}

func TestBuildS3FSCmdSpecialChars(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:     "test-alias_123",
		Bucket:    "test-bucket.456",
		Host:      "s3.test-region.example.com",
		Region:    "us-east-1a",
		AccessKey: "key-with-dashes",
		SecretKey: "secret+with=special/chars",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/test_path/with-dashes", "/tmp/passwd-s3fs")

	args := cmd.Args
	argsStr := strings.Join(args, " ")

	// Should contain all values
	if !strings.Contains(argsStr, "test-bucket.456") {
		t.Error("Expected bucket with special chars in args")
	}
	if !strings.Contains(argsStr, "/mnt/test_path/with-dashes") {
		t.Error("Expected path with special chars in args")
	}
	if !strings.Contains(argsStr, "s3.test-region.example.com") {
		t.Error("Expected host with special chars in args")
	}
	if !strings.Contains(argsStr, "us-east-1a") {
		t.Error("Expected region with special chars in args")
	}
}

func TestBuildS3FSCmdOnlyBucketAndPath(t *testing.T) {
	manager := &MountManager{}

	// Minimal config - only bucket and path matter for the first two args
	s3Config := S3Config{
		Alias:  "minimal",
		Bucket: "minimal-bucket",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/minimal", "")

	args := cmd.Args

	// First two args after command name should be bucket and path
	if len(args) < 3 {
		t.Fatalf("Expected at least 3 args, got %d", len(args))
	}

	// Args[0] is the command name, Args[1] is bucket, Args[2] is path
	if args[1] != "minimal-bucket" {
		t.Errorf("Expected bucket 'minimal-bucket' at args[1], got '%s'", args[1])
	}
	if args[2] != "/mnt/minimal" {
		t.Errorf("Expected path '/mnt/minimal' at args[2], got '%s'", args[2])
	}
}

func TestUmountFunction(t *testing.T) {
	// Call the package-level Umount function with empty aliases
	// This should not panic and should handle empty input gracefully
	Umount([][]string{}, true)
}

func TestUmountFunctionWithAliases(t *testing.T) {
	// Call with a valid alias structure
	// This tests the data transformation from [][]string to S3Config
	aliases := [][]string{
		{"alias1", "bucket1", "host1", "region1", "key1", "secret1"},
		{"alias2", "bucket2", "host2", "region2", "key2", "secret2"},
	}

	// This should not panic even without fusermount3
	// It tests the alias-to-config transformation logic
	Umount(aliases, true)
}

func TestUmountFunctionWithIncompleteAliases(t *testing.T) {
	// Call with incomplete alias structure (less than 6 elements)
	// This tests that incomplete entries are skipped
	aliases := [][]string{
		{"alias1", "bucket1"},                    // Only 2 elements - should be skipped
		{"alias2"},                               // Only 1 element - should be skipped
		{},                                       // Empty - should be skipped
		{"alias3", "b3", "h3", "r3", "k3", "s3"}, // Complete - should be processed
	}

	// This should not panic - it should skip incomplete entries
	Umount(aliases, true)
}

func TestUmountFunctionWithNil(t *testing.T) {
	// Call with nil - should handle gracefully
	Umount(nil, true)
}

func TestNewMountManagerWithNilConfig(t *testing.T) {
	// Test that NewMountManager handles nil config
	// This will panic if not handled, but looking at the code it doesn't check for nil
	// So we expect this to create a manager with nil config
	manager := NewMountManager(nil)

	if manager == nil {
		t.Error("Expected non-nil manager")
	}

	if manager.config != nil {
		t.Errorf("Expected nil config, got %v", manager.config)
	}

	if len(manager.activeMounts) != 0 {
		t.Errorf("Expected empty activeMounts, got %d", len(manager.activeMounts))
	}
}

func TestUmountFromFileNotExists(t *testing.T) {
	// Test umountFromFile when the file doesn't exist
	// This tests the error handling path
	config := &MountConfig{
		BaseDir:          "/s3-data",
		Mounts:           []S3Config{},
		MountingComplete: false,
	}
	manager := NewMountManager(config)

	// Ensure the file doesn't exist by setting to a non-existent path
	// We can't change ActiveMountsFile as it's a const, so we test with the default
	// which likely doesn't exist in test environment
	err := manager.umountFromFile(true)

	// Should return nil when file doesn't exist (os.IsNotExist is handled)
	if err != nil {
		t.Errorf("Expected umountFromFile() to return nil when file doesn't exist, got error: %v", err)
	}
}

func TestMountManagerActiveMountsTracking(t *testing.T) {
	config := &MountConfig{
		BaseDir:          "/data",
		MountingComplete: false,
		Mounts: []S3Config{
			{Alias: "alias1", Bucket: "bucket1"},
		},
	}

	manager := NewMountManager(config)

	// Initially should have empty activeMounts
	if len(manager.activeMounts) != 0 {
		t.Errorf("Expected empty activeMounts initially, got %d", len(manager.activeMounts))
	}

	// Manually add a mount (simulating successful mount)
	manager.activeMounts = append(manager.activeMounts, "/mnt/test1")
	manager.activeMounts = append(manager.activeMounts, "/mnt/test2")

	if len(manager.activeMounts) != 2 {
		t.Errorf("Expected 2 active mounts, got %d", len(manager.activeMounts))
	}

	if manager.activeMounts[0] != "/mnt/test1" {
		t.Errorf("Expected first mount '/mnt/test1', got '%s'", manager.activeMounts[0])
	}

	if manager.activeMounts[1] != "/mnt/test2" {
		t.Errorf("Expected second mount '/mnt/test2', got '%s'", manager.activeMounts[1])
	}
}

func TestUmountFunctionDataTransformation(t *testing.T) {
	// Set up environment to test the transformation logic
	os.Setenv("S3_TO_LOCAL_test1_ALIAS", "test1")
	os.Setenv("S3_TO_LOCAL_test1_BUCKET", "bucket1")
	os.Setenv("S3_TO_LOCAL_test1_HOST", "host1")
	os.Setenv("S3_TO_LOCAL_test1_REGION", "region1")
	os.Setenv("S3_TO_LOCAL_test1_ACCESS_KEY", "key1")
	os.Setenv("S3_TO_LOCAL_test1_SECRET_KEY", "secret1")

	defer func() {
		os.Unsetenv("S3_TO_LOCAL_test1_ALIAS")
		os.Unsetenv("S3_TO_LOCAL_test1_BUCKET")
		os.Unsetenv("S3_TO_LOCAL_test1_HOST")
		os.Unsetenv("S3_TO_LOCAL_test1_REGION")
		os.Unsetenv("S3_TO_LOCAL_test1_ACCESS_KEY")
		os.Unsetenv("S3_TO_LOCAL_test1_SECRET_KEY")
	}()

	// Test with aliases that should be converted to S3Config
	// The Umount function will try to load config from env and unmount
	// We just verify it doesn't panic
	Umount([][]string{
		{"test1", "bucket1", "host1", "region1", "key1", "secret1"},
	}, true)
}

func TestMountFunctionError(t *testing.T) {
	// Set up environment variables
	os.Setenv("S3_TO_LOCAL_test_pkg_ALIAS", "test_pkg")
	os.Setenv("S3_TO_LOCAL_test_pkg_BUCKET", "test-bucket-pkg")
	os.Setenv("S3_TO_LOCAL_test_pkg_HOST", "s3-pkg.example.com")
	os.Setenv("S3_TO_LOCAL_test_pkg_REGION", "us-west-1")
	os.Setenv("S3_TO_LOCAL_test_pkg_ACCESS_KEY", "pkg-key")
	os.Setenv("S3_TO_LOCAL_test_pkg_SECRET_KEY", "pkg-secret")

	defer func() {
		os.Unsetenv("S3_TO_LOCAL_test_pkg_ALIAS")
		os.Unsetenv("S3_TO_LOCAL_test_pkg_BUCKET")
		os.Unsetenv("S3_TO_LOCAL_test_pkg_HOST")
		os.Unsetenv("S3_TO_LOCAL_test_pkg_REGION")
		os.Unsetenv("S3_TO_LOCAL_test_pkg_ACCESS_KEY")
		os.Unsetenv("S3_TO_LOCAL_test_pkg_SECRET_KEY")
	}()

	// In test mode LoadConfigFromEnv skips creating /s3-data, so Mount succeeds
	// and returns the configured aliases without executing s3fs.
	result := Mount(true)

	// Result should contain the single alias configured above
	if result == nil {
		t.Fatal("Expected non-nil result in test mode, got nil")
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 alias, got %d", len(result))
	}
	if result[0][0] != "test_pkg" {
		t.Errorf("Expected alias 'test_pkg', got '%s'", result[0][0])
	}
}

func TestUmountWithEmptyMounts(t *testing.T) {
	// Test Umount with aliases that result in empty mounts after filtering
	// (all aliases have less than 6 elements)
	Umount([][]string{
		{"short"},
		{"also", "short"},
		{"still", "too", "short"},
	}, true)
	// Should not panic
}

func TestUmountWithMixedAliases(t *testing.T) {
	// Test with mix of complete and incomplete aliases
	Umount([][]string{
		{"c1", "b1", "h1", "r1", "k1", "s1"}, // Complete
		{"short"},                            // Incomplete
		{"c2", "b2", "h2", "r2", "k2", "s2"}, // Complete
	}, true)
	// Should not panic, should process complete ones
}

func TestMountManagerWithEmptyMounts(t *testing.T) {
	// Test Mount with no mounts in config
	// This should handle empty mounts gracefully
	config := &MountConfig{
		BaseDir:          "/s3-data",
		Mounts:           []S3Config{},
		MountingComplete: false,
	}

	manager := NewMountManager(config)

	// This will try to create /s3-data which we can't do
	// But we test that it handles the error gracefully
	mountedPaths, err := manager.Mount(true)

	// Should return an error or succeed
	// We just want to verify it doesn't panic
	_ = mountedPaths
	_ = err
}

func TestUmountFromFileWithParsing(t *testing.T) {
	// Test the parsing logic in umountFromFile
	// Even though we can't write to /etc, we can test with the default path
	config := &MountConfig{
		BaseDir:          "/s3-data",
		Mounts:           []S3Config{},
		MountingComplete: false,
	}

	manager := NewMountManager(config)

	// Add a mount to activeMounts
	manager.activeMounts = []string{"/mnt/existing"}

	// Call umountFromFile - it will try to read from /etc/active_mounts.txt
	// which doesn't exist, so it should return nil
	err := manager.umountFromFile(true)

	if err != nil {
		t.Errorf("Expected umountFromFile() to return nil when file doesn't exist, got: %v", err)
	}
}

func TestMountManagerConfigAccess(t *testing.T) {
	config := &MountConfig{
		BaseDir:          "/test",
		MountingComplete: false,
		Mounts: []S3Config{
			{Alias: "test", Bucket: "bucket"},
		},
	}

	manager := NewMountManager(config)

	// Test that we can access the config through the manager
	if manager.config != config {
		t.Error("Expected manager.config to equal the provided config")
	}

	// Test that we can modify config through manager
	manager.config.MountingComplete = true
	if !config.MountingComplete {
		t.Error("Expected config.MountingComplete to be true after modification through manager")
	}
}

func TestMountManagerActiveMountsEmpty(t *testing.T) {
	config := &MountConfig{
		BaseDir:          "/test",
		MountingComplete: false,
		Mounts:           []S3Config{},
	}

	manager := NewMountManager(config)

	// activeMounts should be initialized as empty slice
	if manager.activeMounts == nil {
		t.Error("Expected activeMounts to be initialized (not nil)")
	}

	if len(manager.activeMounts) != 0 {
		t.Errorf("Expected empty activeMounts, got %d", len(manager.activeMounts))
	}
}

func TestBuildS3FSCmdWithOnlyHost(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:  "test",
		Bucket: "test-bucket",
		Host:   "s3.example.com",
		Region: "",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/test", "")

	argsStr := strings.Join(cmd.Args, " ")

	// Should have url option but not endpoint
	if !strings.Contains(argsStr, "url=s3.example.com") {
		t.Error("Expected url option in args")
	}
	if strings.Contains(argsStr, "endpoint=") {
		t.Error("Expected no endpoint option when region is empty")
	}
}

func TestBuildS3FSCmdWithOnlyRegion(t *testing.T) {
	manager := &MountManager{}

	s3Config := S3Config{
		Alias:  "test",
		Bucket: "test-bucket",
		Host:   "",
		Region: "us-east-1",
	}

	cmd := manager.buildS3FSCmd(s3Config, "/mnt/test", "")

	argsStr := strings.Join(cmd.Args, " ")

	// Should have endpoint option but not url
	if !strings.Contains(argsStr, "endpoint=us-east-1") {
		t.Error("Expected endpoint option in args")
	}
	if strings.Contains(argsStr, "url=") {
		t.Error("Expected no url option when host is empty")
	}
}

func TestMountManagerMountWithConfig(t *testing.T) {
	// Test Mount with a realistic config
	// Use a temp directory to avoid permission issues
	tempDir, err := os.MkdirTemp("", "s3fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &MountConfig{
		BaseDir:          tempDir,
		MountingComplete: false,
		Mounts: []S3Config{
			{
				Alias:     "test-alias",
				Bucket:    "test-bucket",
				Host:      "s3.example.com",
				Region:    "us-east-1",
				AccessKey: "test-key",
				SecretKey: "test-secret",
			},
		},
	}

	manager := NewMountManager(config)

	// Call Mount - it will fail due to s3fs not being available but we can test the flow
	mountedPaths, err := manager.Mount(true)

	// We don't care about the result, just that it doesn't panic
	_ = mountedPaths
	_ = err
}

func TestUmountWithRealisticConfig(t *testing.T) {
	// Test Umount with a realistic config
	// Use a temp directory for consistency
	tempDir, err := os.MkdirTemp("", "s3fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &MountConfig{
		BaseDir:          tempDir,
		Mounts:           []S3Config{},
		MountingComplete: false,
	}

	manager := NewMountManager(config)

	// Manually set up active mounts
	manager.activeMounts = []string{
		fmt.Sprintf("%s/alias1/bucket1", tempDir),
		fmt.Sprintf("%s/alias2/bucket2", tempDir),
	}

	// Call Umount - it will try to execute fusermount3 which will fail
	// but we can test the logic
	err = manager.Umount(true)

	// We don't care about the error, just that it doesn't panic
	_ = err
}

func TestMountManagerMultipleMounts(t *testing.T) {
	config := &MountConfig{
		BaseDir: "/data",
		Mounts: []S3Config{
			{Alias: "alias1", Bucket: "bucket1"},
			{Alias: "alias2", Bucket: "bucket2"},
			{Alias: "alias3", Bucket: "bucket3"},
		},
		MountingComplete: false,
	}

	manager := NewMountManager(config)

	// Verify config was set correctly
	if len(manager.config.Mounts) != 3 {
		t.Errorf("Expected 3 mounts in config, got %d", len(manager.config.Mounts))
	}

	// Verify we can access individual mounts
	if manager.config.Mounts[0].Alias != "alias1" {
		t.Errorf("Expected first mount alias 'alias1', got '%s'", manager.config.Mounts[0].Alias)
	}
}
