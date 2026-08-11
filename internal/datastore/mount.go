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
	"os/exec"
	"strings"
)

// MountManager manages S3FS mounts
type MountManager struct {
	config *MountConfig
	// activeMounts tracks successfully mounted paths for cleanup
	activeMounts []string
}

// NewMountManager creates a new MountManager with the given configuration
func NewMountManager(config *MountConfig) *MountManager {
	return &MountManager{
		config:       config,
		activeMounts: []string{},
	}
}

// Mount performs the S3FS mounting for all configured S3 buckets
// Returns the list of mounted paths or an error
func (m *MountManager) Mount() ([]string, error) {
	fmt.Println("Initializing S3 mounts...")

	// Create folders for each mount
	if err := m.createFolders(); err != nil {
		return nil, fmt.Errorf("failed to create folders: %w", err)
	}

	// Perform S3FS mounts
	if err := m.createS3Mounts(); err != nil {
		return nil, fmt.Errorf("failed to create S3 mounts: %w", err)
	}

	if len(m.config.Mounts) == 0 {
		fmt.Println("No environment variables matching 'S3_TO_LOCAL_*_ALIAS' found.")
	} else {
		fmt.Printf("Processed %d S3 alias(es).\n", len(m.config.Mounts))
	}

	m.config.MountingComplete = true
	fmt.Println("S3 system ready.")

	return m.activeMounts, nil
}

// createFolders creates the target directories for all mounts
func (m *MountManager) createFolders() error {
	for _, s3Config := range m.config.Mounts {
		targetPath := s3Config.GetMountPath(m.config.BaseDir)
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", targetPath, err)
		}
		fmt.Printf("Created folder: %s\n", targetPath)
	}
	return nil
}

// createS3Mounts performs the actual S3FS mount for each configuration
func (m *MountManager) createS3Mounts() error {
	for _, s3Config := range m.config.Mounts {
		targetPath := s3Config.GetMountPath(m.config.BaseDir)
		
		// Create password file for S3FS authentication
		passwdFile := "/tmp/passwd-s3fs"
		if err := m.createPasswordFile(passwdFile, s3Config.AccessKey, s3Config.SecretKey); err != nil {
			return fmt.Errorf("failed to create password file: %w", err)
		}
		defer os.Remove(passwdFile)

		// Build S3FS command
		cmd := m.buildS3FSCmd(s3Config, targetPath, passwdFile)
		
		// Execute S3FS mount
		fmt.Printf("Mounting S3 bucket %s to %s\n", s3Config.Bucket, targetPath)
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Printf("s3fs returned a non-zero status (%d) for alias '%s'\n",
					exitErr.ExitCode(), s3Config.Alias)
				// Continue to next mount even if this one fails
				continue
			}
			return fmt.Errorf("failed to execute s3fs: %w", err)
		}

		fmt.Printf("Successfully mounted '%s'\n", s3Config.Alias)
		m.activeMounts = append(m.activeMounts, targetPath)
		
		// Write to active mounts file for tracking
		if err := m.writeActiveMount(targetPath); err != nil {
			fmt.Printf("Warning: failed to write active mount: %v\n", err)
		}
	}
	
	return nil
}

// createPasswordFile creates the password file for S3FS authentication
func (m *MountManager) createPasswordFile(path, accessKey, secretKey string) error {
	content := fmt.Sprintf("%s:%s\n", accessKey, secretKey)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write password file: %w", err)
	}
	return nil
}

// buildS3FSCmd builds the S3FS command for mounting
func (m *MountManager) buildS3FSCmd(s3Config S3Config, targetPath, passwdFile string) *exec.Cmd {
	args := []string{
		s3Config.Bucket,
		targetPath,
	}
	
	// Add options - only include non-empty values
	if passwdFile != "" {
		args = append(args, "-o", fmt.Sprintf("passwd_file=%s", passwdFile))
	}
	if s3Config.Host != "" {
		args = append(args, "-o", fmt.Sprintf("url=%s", s3Config.Host))
	}
	if s3Config.Region != "" {
		args = append(args, "-o", fmt.Sprintf("endpoint=%s", s3Config.Region))
	}
	
	// Always add these flags
	args = append(args,
		"-o", "use_path_request_style",
		"-o", "allow_other",
		"-o", "nonempty",
		"-f", // Run in foreground (for container)
	)
	
	// Build the command - use s3fs binary
	// Note: In the Dockerfile, s3fs is installed from debian package
	return exec.Command("s3fs", args...)
}

// writeActiveMount writes a mounted path to the active mounts tracking file
func (m *MountManager) writeActiveMount(path string) error {
	f, err := os.OpenFile(ActiveMountsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	if _, err := f.WriteString(path + "\n"); err != nil {
		return err
	}
	return nil
}

// Umount unmounts all active S3FS mounts
func (m *MountManager) Umount() error {
	fmt.Println("Starting unmount sequence...")
	
	for _, mountPath := range m.activeMounts {
		fmt.Printf("Unmounting: %s\n", mountPath)
		
		// Use fusermount3 to unmount (available in the Docker image)
		cmd := exec.Command("fusermount3", "-u", mountPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to unmount %s. It might be busy.\n", mountPath)
			// Continue trying other mounts
			continue
		}
		fmt.Printf("Successfully unmounted %s\n", mountPath)
	}
	
	// Also try reading from active mounts file for any mounts we might have missed
	if err := m.umountFromFile(); err != nil {
		fmt.Printf("Warning: Failed to unmount from file: %v\n", err)
	}
	
	return nil
}

// umountFromFile reads the active mounts file and unmounts each path
func (m *MountManager) umountFromFile() error {
	content, err := os.ReadFile(ActiveMountsFile)
	if err != nil {
		// File might not exist, which is fine
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check if we already unmounted this
		alreadyUnmounted := false
		for _, path := range m.activeMounts {
			if path == line {
				alreadyUnmounted = true
				break
			}
		}
		
		if !alreadyUnmounted {
			fmt.Printf("Unmounting (from file): %s\n", line)
			cmd := exec.Command("fusermount3", "-u", line)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: Failed to unmount %s. It might be busy.\n", line)
			}
		}
	}
	
	return nil
}

// Mount performs the mounting and returns the configuration with aliases
// This function matches the Python mount.mount() signature
func Mount() [][]string {
	config, err := LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return nil
	}
	
	manager := NewMountManager(config)
	_, err = manager.Mount()
	if err != nil {
		fmt.Printf("Error mounting: %v\n", err)
		return nil
	}
	
	// Convert to the format expected by Python (list of lists)
	// Python returns: [alias, bucket, host, region, access_key, secret_key]
	var result [][]string
	for _, s3Config := range config.Mounts {
		result = append(result, []string{
			s3Config.Alias,
			s3Config.Bucket,
			s3Config.Host,
			s3Config.Region,
			s3Config.AccessKey,
			s3Config.SecretKey,
		})
	}
	
	return result
}

// Umount unmounts all mounts (matches Python mount.umount() signature)
// In Python: mount.umount(aliases) where aliases is list of lists
func Umount(aliases [][]string) {
	config := &MountConfig{
		BaseDir:   DefaultBaseDir,
		Mounts:    []S3Config{},
		MountingComplete: false,
	}
	
	// Convert Python-style aliases to S3Config
	for _, alias := range aliases {
		if len(alias) >= 6 {
			config.Mounts = append(config.Mounts, S3Config{
				Alias:     alias[0],
				Bucket:    alias[1],
				Host:      alias[2],
				Region:    alias[3],
				AccessKey: alias[4],
				SecretKey: alias[5],
			})
		}
	}
	
	manager := NewMountManager(config)
	// Build active mounts from config
	for _, s3Config := range config.Mounts {
		manager.activeMounts = append(manager.activeMounts, s3Config.GetMountPath(config.BaseDir))
	}
	
	manager.Umount()
}
