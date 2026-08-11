/*
This file is part of REANA.
Copyright (C) 2026 CERN.

REANA is free software; you can redistribute it and/or modify it
under the terms of the MIT License; see LICENSE file for more details.
*/

package version

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("version must not be empty")
	}
	// Test GetVersion function for coverage
	if GetVersion() != Version {
		t.Errorf("GetVersion() should return Version, got %s", GetVersion())
	}
}
