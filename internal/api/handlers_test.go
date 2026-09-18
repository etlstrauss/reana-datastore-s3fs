/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

// Package api provides endpoints for simple  functionalities like health check or others
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/reanahub/reana-datastore-s3fs/internal/datastore"
)

// Test the health endpoint of the handler when mounting is done
func TestHealthHandlerReady(t *testing.T) {
	config := &datastore.MountConfig{
		BaseDir:          "/test",
		Mounts:           []datastore.S3Config{},
		MountingComplete: true,
	}

	server := &Server{
		Config: config,
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", response.Status)
	}
}

// REst the health endpoint of the handler when mounting is false
func TestHealthHandlerMounting(t *testing.T) {
	config := &datastore.MountConfig{
		BaseDir:          "/test",
		Mounts:           []datastore.S3Config{},
		MountingComplete: false,
	}

	server := &Server{
		Config: config,
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.HealthHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "mounting" {
		t.Errorf("Expected status 'mounting', got '%s'", response.Status)
	}
}

// Test shutdown method of api (Post)
func TestShutdownHandlerPost(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodPost, "/shutdown", nil)
	w := httptest.NewRecorder()

	server.ShutdownHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "Cleanup initiated" {
		t.Errorf("Expected body 'Cleanup initiated', got '%s'", body)
	}
}

// Test shutdown method of api (Get)
func TestShutdownHandlerGet(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/shutdown", nil)
	w := httptest.NewRecorder()

	server.ShutdownHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// Test the NewServer object on functionality
func TestNewServer(t *testing.T) {
	// Clear environment
	os.Setenv("S3_TO_LOCAL_", "")

	server := NewServer(true)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.Config == nil {
		t.Error("Expected server.Config to be set")
	}

	if server.MountManager == nil {
		t.Error("Expected server.MountManager to be set")
	}

	if len(server.Aliases) != 0 {
		t.Errorf("Expected empty Aliases, got %d", len(server.Aliases))
	}
}
