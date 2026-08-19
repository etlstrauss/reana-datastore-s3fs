/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/reanahub/reana-datastore-s3fs/internal/api"
	"github.com/reanahub/reana-datastore-s3fs/internal/datastore"
	"github.com/reanahub/reana-datastore-s3fs/internal/version"
)

var (
	// aliases is used to track mounted aliases for cleanup (global for signal handler)
	aliases [][]string
	// mountManager is used for cleanup on signal
	mountManager *datastore.MountManager
	// config is the loaded configuration
	config *datastore.MountConfig
)

func main() {
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		return
	}

	// Load configuration
	var err error
	config, err = datastore.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create mount manager
	mountManager = datastore.NewMountManager(config)

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start cleanup in a goroutine when signal is received
	go func() {
		<-signalChan
		cleanupAndExit()
	}()

	// Perform initial mounting
	fmt.Println("Initializing S3 mounts...")
	_, err = mountManager.Mount()
	if err != nil {
		fmt.Printf("Error during mounting: %v\n", err)
		// Don't exit - we still want to start the server
	}

	// Store aliases for cleanup
	for _, s3Config := range config.Mounts {
		aliases = append(aliases, []string{
			s3Config.Alias,
			s3Config.Bucket,
			s3Config.Host,
			s3Config.Region,
			s3Config.AccessKey,
			s3Config.SecretKey,
		})
	}

	config.MountingComplete = true
	fmt.Println("S3 system ready.")

	// Start the API server
	server := api.NewServer()
	if err := server.Run(); err != nil {
		fmt.Printf("Error running server: %v\n", err)
		os.Exit(1)
	}
}

func cleanupAndExit() {
	fmt.Println("\nSignal received. Cleaning up mounts...")

	if mountManager != nil {
		mountManager.Umount()
	} else {
		// Fallback: use the Python-compatible Umount function
		datastore.Umount(aliases)
	}

	fmt.Println("Exiting.")
	os.Exit(0)
}
