/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package main

import (
	"flag"
	"os"
	"testing"
)

func TestMainPackage(t *testing.T) {
	// Test that the package can be imported and cleanupAndExit works
	// This ensures the cmd package has coverage

	// Save original exitFunc
	originalExit := exitFunc
	defer func() { exitFunc = originalExit }()

	// Override exitFunc to not actually exit
	exitFunc = func(code int) {}

	// Initialize globals to nil/empty to test the cleanup path
	config = nil
	mountManager = nil
	aliases = [][]string{}

	// Now we can call cleanupAndExit without it exiting
	cleanupAndExit()
}

func TestMainVersionFlag(t *testing.T) {
	// Test the version flag path in main()

	// Save original exitFunc
	originalExit := exitFunc
	defer func() { exitFunc = originalExit }()

	// Override exitFunc to not actually exit
	exitFunc = func(code int) {}

	// Save and restore original os.Args and flag.CommandLine
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up args for version flag
	os.Args = []string{"reana-datastore-s3fs", "-version"}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Run main - it should print version and call exitFunc(0)
	main()
}
