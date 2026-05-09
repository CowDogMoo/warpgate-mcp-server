// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package prompts

import (
	"os"
	"strings"
	"testing"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/server"
)

func createTestLogger(t *testing.T) (*logging.Logger, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "warpgate-test-log-*")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v", err)
	}
	_ = tmpFile.Close()

	logger, err := logging.NewLogger(tmpFile.Name())
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create logger: %v", err)
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	return logger, cleanup
}

func newTestServer() *server.MCPServer {
	return server.NewMCPServer("test", "1.0.0", server.WithPromptCapabilities(true))
}

func TestRegisterPrompts(t *testing.T) {
	s := newTestServer()
	logger, cleanup := createTestLogger(t)
	defer cleanup()

	RegisterPrompts(s, logger)
	t.Log("RegisterPrompts completed without panic")
}

func TestRegisterPromptsMultipleTimes(t *testing.T) {
	logger, cleanup := createTestLogger(t)
	defer cleanup()

	// Re-registering on a fresh server should be safe; mcp-go panics if the
	// same prompt name is added twice on a single server instance, which is
	// expected behavior, so each iteration uses a new server.
	for i := 0; i < 3; i++ {
		s := newTestServer()
		RegisterPrompts(s, logger)
	}
	t.Log("RegisterPrompts on fresh servers completed without panic")
}

func TestRequireArg(t *testing.T) {
	cases := []struct {
		name    string
		args    map[string]string
		key     string
		want    string
		wantErr bool
	}{
		{name: "present", args: map[string]string{"k": "v"}, key: "k", want: "v"},
		{name: "trims whitespace", args: map[string]string{"k": "  v  "}, key: "k", want: "v"},
		{name: "missing", args: map[string]string{}, key: "k", wantErr: true},
		{name: "empty", args: map[string]string{"k": ""}, key: "k", wantErr: true},
		{name: "whitespace only", args: map[string]string{"k": "   "}, key: "k", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := requireArg(tc.args, tc.key)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func assertContainsAll(t *testing.T, text string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(text, s) {
			t.Errorf("expected rendered prompt to contain %q\n--- rendered ---\n%s\n--- end ---", s, text)
		}
	}
}

func TestRenderBootstrapNewTemplate(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		got := renderBootstrapNewTemplate("attack-box", "", "")
		assertContainsAll(t, got,
			`"attack-box"`,
			`name="attack-box"`,
			"warpgate_init",
			"warpgate_validate",
			"warpgate_schema_validate",
			"warpgate_build_streaming",
			"warpgate://schema",
		)
		if strings.Contains(got, "Fork from") {
			t.Errorf("did not expect Fork-from line when from is empty:\n%s", got)
		}
	})

	t.Run("with from and output", func(t *testing.T) {
		got := renderBootstrapNewTemplate("attack-box", "kali-base", "/tmp/out")
		assertContainsAll(t, got,
			`Fork from existing template "kali-base"`,
			`from="kali-base"`,
			`output="/tmp/out"`,
		)
	})
}

func TestRenderDebugFailedBuild(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		got := renderDebugFailedBuild("./blueprints/attack-box", "")
		assertContainsAll(t, got,
			"./blueprints/attack-box",
			"warpgate_validate",
			"warpgate_schema_validate",
			"warpgate://schema",
			"warpgate://cli/info",
			"warpgate://config",
		)
		if strings.Contains(got, "Reported failure:") {
			t.Errorf("did not expect failure line when summary is empty:\n%s", got)
		}
	})

	t.Run("with summary", func(t *testing.T) {
		got := renderDebugFailedBuild("./bp", "exit 127 in install step")
		assertContainsAll(t, got, "Reported failure: exit 127 in install step")
	})
}

func TestRenderAddProvisioner(t *testing.T) {
	t.Run("shell with description", func(t *testing.T) {
		got := renderAddProvisioner("./bp/attack-box", "shell", "install nmap and masscan")
		assertContainsAll(t, got,
			"shell provisioner",
			"./bp/attack-box",
			"Goal: install nmap and masscan",
			"./bp/attack-box/scripts/",
			"warpgate://schema",
			"warpgate_validate",
		)
	})

	t.Run("ansible without description", func(t *testing.T) {
		got := renderAddProvisioner("./bp/x", "ansible", "")
		assertContainsAll(t, got, "ansible provisioner")
		if strings.Contains(got, "Goal:") {
			t.Errorf("did not expect Goal: line when description is empty:\n%s", got)
		}
	})
}

func TestRenderConvertFromPacker(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		got := renderConvertFromPacker("./packer/attack.pkr.hcl", "")
		assertContainsAll(t, got,
			"./packer/attack.pkr.hcl",
			`packer_path="./packer/attack.pkr.hcl"`,
			"warpgate_convert",
			"warpgate_validate",
			"warpgate_build_streaming",
		)
		if strings.Contains(got, ", output=") {
			t.Errorf("did not expect output kwarg when output_dir is empty:\n%s", got)
		}
	})

	t.Run("with output_dir", func(t *testing.T) {
		got := renderConvertFromPacker("./packer/x.pkr.hcl", "./blueprints/x")
		assertContainsAll(t, got, `output="./blueprints/x"`)
	})
}

func TestRenderPublishMultiarchImage(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		got := renderPublishMultiarchImage("attack-box", "ghcr.io/cowdogmoo", "", "")
		assertContainsAll(t, got,
			"ghcr.io/cowdogmoo/attack-box:latest",
			"amd64,arm64",
			"warpgate_build_streaming",
			"warpgate_manifests_create",
			"warpgate_manifests_push",
			"warpgate_registry_inspect",
			"warpgate://config",
		)
	})

	t.Run("trims trailing slash on registry", func(t *testing.T) {
		got := renderPublishMultiarchImage("attack-box", "ghcr.io/cowdogmoo/", "v1.2.3", "amd64")
		assertContainsAll(t, got,
			"ghcr.io/cowdogmoo/attack-box:v1.2.3",
		)
		if strings.Contains(got, "cowdogmoo//attack-box") {
			t.Errorf("trailing slash on registry was not trimmed:\n%s", got)
		}
	})
}
