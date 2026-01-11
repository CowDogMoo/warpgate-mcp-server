// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package version

import (
	"strings"
	"testing"
)

func TestVersionVariables(t *testing.T) {
	// Version should be set
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// GitCommit defaults to "unknown"
	if GitCommit == "" {
		t.Error("GitCommit should not be empty (defaults to 'unknown')")
	}

	// BuildDate defaults to "unknown"
	if BuildDate == "" {
		t.Error("BuildDate should not be empty (defaults to 'unknown')")
	}
}

func TestGetHumanVersion(t *testing.T) {
	humanVersion := GetHumanVersion()

	// Should start with "v"
	if !strings.HasPrefix(humanVersion, "v") {
		t.Errorf("GetHumanVersion() = %q, should start with 'v'", humanVersion)
	}

	// Should contain the version
	if !strings.Contains(humanVersion, Version) {
		t.Errorf("GetHumanVersion() = %q, should contain Version %q", humanVersion, Version)
	}

	// Expected format: vX.Y.Z
	expected := "v" + Version
	if humanVersion != expected {
		t.Errorf("GetHumanVersion() = %q, want %q", humanVersion, expected)
	}
}

func TestGetFullVersion(t *testing.T) {
	fullVersion := GetFullVersion()

	// Should contain version info
	if !strings.Contains(fullVersion, "Version:") {
		t.Errorf("GetFullVersion() missing 'Version:' marker")
	}

	// Should contain commit info
	if !strings.Contains(fullVersion, "Commit:") {
		t.Errorf("GetFullVersion() missing 'Commit:' marker")
	}

	// Should contain build date info
	if !strings.Contains(fullVersion, "Build Date:") {
		t.Errorf("GetFullVersion() missing 'Build Date:' marker")
	}

	// Should contain the actual version
	if !strings.Contains(fullVersion, Version) {
		t.Errorf("GetFullVersion() should contain Version %q", Version)
	}

	// Should contain git commit
	if !strings.Contains(fullVersion, GitCommit) {
		t.Errorf("GetFullVersion() should contain GitCommit %q", GitCommit)
	}

	// Should contain build date
	if !strings.Contains(fullVersion, BuildDate) {
		t.Errorf("GetFullVersion() should contain BuildDate %q", BuildDate)
	}
}

func TestVersionFormat(t *testing.T) {
	// Test that version is in expected semver format
	parts := strings.Split(Version, ".")
	if len(parts) != 3 {
		t.Errorf("Version %q should be in semver format (X.Y.Z)", Version)
	}
}
