/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package datastore

import (
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
}

func TestCreatePasswordFile(t *testing.T) {
	manager := &MountManager{}

	tmpFile := "/tmp/test_passwd_s3fs_" + string(rune(os.Getpid()))
	defer os.Remove(tmpFile)

	err := manager.createPasswordFile(tmpFile, "access-key-123", "secret-key-456")
	if err != nil {
		t.Fatalf("createPasswordFile() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Password file was not created")
	}

	// Check file permissions (should be 0600)
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat password file: %v", err)
	}

	// Check permissions (0600 = 0o600)
	mode := info.Mode()
	if mode.Perm() != 0o600 {
		t.Errorf("Expected permissions 0600, got %o", mode.Perm())
	}

	// Check file content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read password file: %v", err)
	}

	expectedContent := "access-key-123:secret-key-456\n"
	if string(content) != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, string(content))
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
