// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

// knownCLISubcommands is the audited per-subcommand flag set. Flag names use
// the kebab-case spelling that `warpgate <sub> --help` prints. When the
// parity test fails for a new CLI flag, either expose it via MCP and add it
// here, or add it to intentionallySkipped with a rationale.
var knownCLISubcommands = map[string]map[string]bool{
	"build": {
		"template": true, "from-git": true, "target": true, "arch": true,
		"push": true, "push-digest": true, "registry": true, "var": true,
		"var-file": true, "build-arg": true, "tag": true, "label": true,
		"cache-from": true, "cache-to": true, "no-cache": true,
		"save-digests": true, "digest-dir": true,
		"region": true, "instance-type": true, "force": true, "cleanup": true,
		"dry-run": true, "regions": true, "parallel-regions": true,
		"copy-to-regions": true, "stream-logs": true, "show-ec2-status": true,
		"output-manifest": true,
		"subscription":    true, "location": true, "resource-group": true,
		"gallery": true, "image-definition": true, "vm-size": true,
		"identity-id": true, "target-regions": true, "subnet-id": true,
		"proxy-vm-size":    true,
		"proxmox-endpoint": true, "proxmox-node": true, "proxmox-storage": true,
		"proxmox-pool": true,
	},
	"validate": {
		"syntax-only": true,
	},
	"init": {
		"from": true, "output": true,
	},
	"templates list": {
		"format": true, "source": true, "quiet": true,
	},
	"templates info":   {},
	"templates add":    {},
	"templates remove": {},
	"templates search": {},
	"templates update": {},
	"manifests create": {
		"name": true, "registry": true, "namespace": true, "auth-file": true,
		"tag": true, "digest-dir": true, "verify-registry": true,
		"verify-concurrency": true, "max-age": true, "require-arch": true,
		"best-effort": true, "annotation": true, "label": true,
		"health-check": true, "show-diff": true, "no-progress": true,
		"dry-run": true, "force": true, "quiet": true, "verbose": true,
	},
	"manifests list": {
		"name": true, "registry": true, "namespace": true, "auth-file": true,
	},
	"manifests inspect": {
		"name": true, "registry": true, "namespace": true, "auth-file": true,
		"tag": true,
	},
	"convert packer": {
		"author": true, "license": true, "version": true, "base-image": true,
		"include-ami": true, "output": true, "dry-run": true,
	},
	"config get":  {},
	"config set":  {},
	"config show": {},
	"config init": {"force": true},
	"config path": {},
	"cleanup": {
		"region": true, "dry-run": true, "all": true, "versions": true,
		"keep": true, "yes": true,
	},
}

// intentionallySkipped lists CLI flags that exist in `warpgate` but are
// deliberately not exposed through MCP. Each entry must have a rationale in a
// comment.
var intentionallySkipped = map[string]map[string]bool{}
