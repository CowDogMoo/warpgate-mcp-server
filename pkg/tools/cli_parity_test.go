// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestCLIParity checks each warpgate CLI subcommand's flag set against
// cli_flag_inventory.go. The error messages tell the developer what to change
// when a mismatch fires. Skipped when the `warpgate` binary isn't on PATH.
func TestCLIParity(t *testing.T) {
	if _, err := exec.LookPath("warpgate"); err != nil {
		t.Skip("warpgate binary not on PATH; skipping CLI parity check")
	}

	subcommands := make([]string, 0, len(knownCLISubcommands))
	for sub := range knownCLISubcommands {
		subcommands = append(subcommands, sub)
	}
	sort.Strings(subcommands)

	for _, sub := range subcommands {
		t.Run(sub, func(t *testing.T) {
			actual := extractCLIFlags(t, sub)

			known := knownCLISubcommands[sub]
			skipped := intentionallySkipped[sub]

			// Direction 1: CLI flag → must be known or skipped.
			for flag := range actual {
				if known[flag] || skipped[flag] {
					continue
				}
				t.Errorf("warpgate %s: new CLI flag --%s is not in the MCP inventory.\n"+
					"  Action: expose it as an MCP parameter and add to knownCLISubcommands[%q],\n"+
					"          or add it to intentionallySkipped[%q] with a rationale.",
					sub, flag, sub, sub)
			}

			// Direction 2: inventory flag → must exist in the CLI.
			for flag := range known {
				if !actual[flag] {
					t.Errorf("warpgate %s: inventory flag --%s no longer appears in CLI help.\n"+
						"  Action: remove from knownCLISubcommands[%q] (and the MCP schema if exposed).",
						sub, flag, sub)
				}
			}
			for flag := range skipped {
				if !actual[flag] {
					t.Errorf("warpgate %s: intentionally-skipped flag --%s no longer appears in CLI help.\n"+
						"  Action: remove from intentionallySkipped[%q].",
						sub, flag, sub)
				}
			}
		})
	}
}

// flagLinePattern matches the start of a cobra flag declaration line:
//
//	"  -f, --foo string   Description"
//	"      --bar          Description"
//
// Group 1 is the long flag name.
var flagLinePattern = regexp.MustCompile(`^\s+(?:-\w,\s+)?--([a-z][a-z0-9-]*)\b`)

// extractCLIFlags returns the per-subcommand flag surface: local Flags plus
// Global Flags inherited from intermediate parents, but not from the root
// (those are universal and aren't part of any subcommand's MCP surface).
func extractCLIFlags(t *testing.T, sub string) map[string]bool {
	t.Helper()

	rootGlobals := rootGlobalsOnce(t)

	out := runHelp(t, sub)
	local := parseFlagSection(out, "Flags:")
	global := parseFlagSection(out, "Global Flags:")

	flags := make(map[string]bool, len(local)+len(global))
	for f := range local {
		flags[f] = true
	}
	for f := range global {
		if !rootGlobals[f] {
			flags[f] = true // parent-persistent, e.g. manifests's --registry
		}
	}
	return flags
}

var rootGlobalsCache map[string]bool

func rootGlobalsOnce(t *testing.T) map[string]bool {
	t.Helper()
	if rootGlobalsCache != nil {
		return rootGlobalsCache
	}
	rootGlobalsCache = parseFlagSection(runHelp(t, ""), "Flags:")
	return rootGlobalsCache
}

func runHelp(t *testing.T, sub string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	args := strings.Fields(sub)
	args = append(args, "--help")
	cmd := exec.CommandContext(ctx, "warpgate", args...) //nolint:gosec // G204: test-only invocation with literal subcommand
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("`warpgate %s --help` failed: %v\nOutput: %s", sub, err, out)
	}
	return string(out)
}

// parseFlagSection returns the long flag names declared under a section
// header ("Flags:" or "Global Flags:") in cobra help output. `--help` is
// excluded.
func parseFlagSection(helpOut, header string) map[string]bool {
	flags := make(map[string]bool)
	inSection := false
	for _, line := range strings.Split(helpOut, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == header {
			inSection = true
			continue
		}
		// Any other section header ends the current section.
		if strings.HasSuffix(trimmed, ":") && trimmed != "" {
			inSection = false
			continue
		}
		if !inSection {
			continue
		}
		if m := flagLinePattern.FindStringSubmatch(line); m != nil {
			if m[1] == "help" {
				continue
			}
			flags[m[1]] = true
		}
	}
	return flags
}
