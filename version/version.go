// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package version

import "fmt"

var (
	// Version is the semantic version of the application
	Version = "0.1.0"

	// GitCommit is the git commit hash (set via ldflags)
	GitCommit = "unknown"

	// BuildDate is the build date (set via ldflags)
	BuildDate = "unknown"
)

// GetHumanVersion returns a human-readable version string
func GetHumanVersion() string {
	return fmt.Sprintf("v%s", Version)
}

// GetFullVersion returns the full version with commit and build date
func GetFullVersion() string {
	return fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s",
		GetHumanVersion(), GitCommit, BuildDate)
}
