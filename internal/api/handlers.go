/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/reanahub/reana-datastore-s3fs/internal/datastore"
)

// HealthResponse represents the JSON response for the health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// Server holds the state of the datastore sidecar server
type Server struct {
	// MountManager is the datastore mount manager
	MountManager *datastore.MountManager
	// Config is the loaded configuration
	Config *datastore.MountConfig
	// Aliases is the list of mounted aliases (for Python compatibility)
	Aliases [][]string
}

// HealthHandler returns 200 if all S3 buckets are ready, 503 otherwise
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if s.Config.MountingComplete {
		response := HealthResponse{Status: "ready"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := HealthResponse{Status: "mounting"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(response)
}

// ShutdownHandler initiates graceful shutdown
func (s *Server) ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Shutdown requested via API endpoint.")

	// Trigger unmount in a goroutine to avoid blocking the response
	go func() {
		if s.MountManager != nil {
			s.MountManager.Umount()
		}
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Cleanup initiated")
}

// NewServer creates a new Server instance
func NewServer() *Server {
	config, err := datastore.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
		config = &datastore.MountConfig{
			BaseDir:          datastore.DefaultBaseDir,
			Mounts:           []datastore.S3Config{},
			MountingComplete: false,
		}
	}

	manager := datastore.NewMountManager(config)

	return &Server{
		MountManager: manager,
		Config:       config,
		Aliases:      [][]string{},
	}
}

// Run starts the HTTP server
func (s *Server) Run() error {
	// Initialize mounts
	fmt.Println("Initializing S3 mounts...")
	if _, err := s.MountManager.Mount(); err == nil {
		// Convert mounts to alias format
		for _, s3Config := range s.Config.Mounts {
			s.Aliases = append(s.Aliases, []string{
				s3Config.Alias,
				s3Config.Bucket,
				s3Config.Host,
				s3Config.Region,
				s3Config.AccessKey,
				s3Config.SecretKey,
			})
		}
	}

	s.Config.MountingComplete = true
	fmt.Println("S3 system ready.")

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.HealthHandler)
	mux.HandleFunc("/shutdown", s.ShutdownHandler)

	fmt.Println("Starting Datastore API on port 5000...")
	return http.ListenAndServe(":5000", mux)
}
