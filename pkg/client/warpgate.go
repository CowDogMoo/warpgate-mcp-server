// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// SourceAll is the sentinel that means "all sources" for `templates list`.
const SourceAll = "all"

// WarpgateClient executes the warpgate CLI.
type WarpgateClient struct {
	binary string
}

// TemplateInfo mirrors the JSON shape emitted by `warpgate templates list --format json`.
// Field tags are PascalCase to match the CLI output.
type TemplateInfo struct {
	Name        string   `json:"Name"`
	Description string   `json:"Description,omitempty"`
	Version     string   `json:"Version,omitempty"`
	Repository  string   `json:"Repository,omitempty"`
	Path        string   `json:"Path,omitempty"`
	Tags        []string `json:"Tags,omitempty"`
	Author      string   `json:"Author,omitempty"`
}

// New returns a client bound to the warpgate binary on PATH (or an explicit path).
func New(binary string) (*WarpgateClient, error) {
	if binary == "" {
		binary = "warpgate"
	}
	if _, err := exec.LookPath(binary); err != nil {
		return nil, fmt.Errorf("warpgate CLI not found (looked for %q): install warpgate from https://github.com/cowdogmoo/warpgate", binary)
	}
	return &WarpgateClient{binary: binary}, nil
}

// Run executes warpgate with the given args, honoring ctx for cancellation, and
// returns combined stdout+stderr. On non-zero exit the output is included in the error.
func (w *WarpgateClient) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, w.binary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("warpgate %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// RunJSON executes warpgate and decodes JSON stdout into v.
func (w *WarpgateClient) RunJSON(ctx context.Context, v interface{}, args ...string) error {
	out, err := w.Run(ctx, args...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(out), v); err != nil {
		return fmt.Errorf("decode JSON from `warpgate %s`: %w\nraw: %s", strings.Join(args, " "), err, out)
	}
	return nil
}

// argBuilder accumulates CLI arguments with helpers for the common flag shapes.
type argBuilder struct{ args []string }

func newArgs(initial ...string) *argBuilder { return &argBuilder{args: initial} }

func (a *argBuilder) raw(parts ...string) *argBuilder {
	a.args = append(a.args, parts...)
	return a
}

// str appends `--flag VALUE` if value is non-empty.
func (a *argBuilder) str(flag, value string) *argBuilder {
	if value != "" {
		a.args = append(a.args, flag, value)
	}
	return a
}

// boolFlag appends `--flag` if v is true.
func (a *argBuilder) boolFlag(flag string, v bool) *argBuilder {
	if v {
		a.args = append(a.args, flag)
	}
	return a
}

// triBool appends `--flag=true|false` if p is non-nil; nothing otherwise.
func (a *argBuilder) triBool(flag string, p *bool) *argBuilder {
	if p != nil {
		a.args = append(a.args, fmt.Sprintf("%s=%t", flag, *p))
	}
	return a
}

// csv appends `--flag a,b,c` (single arg, comma-separated) if values is non-empty.
func (a *argBuilder) csv(flag string, values []string) *argBuilder {
	if len(values) > 0 {
		a.args = append(a.args, flag, strings.Join(values, ","))
	}
	return a
}

// repeated appends `--flag X` once per value.
func (a *argBuilder) repeated(flag string, values []string) *argBuilder {
	for _, v := range values {
		a.args = append(a.args, flag, v)
	}
	return a
}

// kv appends `--flag KEY=VALUE` once per map entry.
func (a *argBuilder) kv(flag string, m map[string]string) *argBuilder {
	for k, v := range m {
		a.args = append(a.args, flag, fmt.Sprintf("%s=%s", k, v))
	}
	return a
}

// intFlag appends `--flag N` if n > 0.
func (a *argBuilder) intFlag(flag string, n int) *argBuilder {
	if n > 0 {
		a.args = append(a.args, flag, fmt.Sprintf("%d", n))
	}
	return a
}

func (a *argBuilder) build() []string { return a.args }

// ListTemplates returns templates from configured sources.
// source may be SourceAll, "local", "git", or a specific repo name; empty means SourceAll.
func (w *WarpgateClient) ListTemplates(ctx context.Context, source string) ([]TemplateInfo, error) {
	args := newArgs("templates", "list", "--format", "json", "--quiet")
	if source != "" && source != SourceAll {
		args.str("--source", source)
	}
	var templates []TemplateInfo
	if err := w.RunJSON(ctx, &templates, args.build()...); err != nil {
		return nil, err
	}
	return templates, nil
}

// SearchTemplates filters the JSON template list locally.
// `warpgate templates search` only emits text, so we list-and-filter to keep the result structured.
func (w *WarpgateClient) SearchTemplates(ctx context.Context, query string) ([]TemplateInfo, error) {
	all, err := w.ListTemplates(ctx, SourceAll)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var matches []TemplateInfo
	for _, t := range all {
		if matchesQuery(t, q) {
			matches = append(matches, t)
		}
	}
	return matches, nil
}

func matchesQuery(t TemplateInfo, q string) bool {
	if strings.Contains(strings.ToLower(t.Name), q) ||
		strings.Contains(strings.ToLower(t.Description), q) ||
		strings.Contains(strings.ToLower(t.Author), q) {
		return true
	}
	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

// TemplateInfoText returns the human-readable info block for a template.
func (w *WarpgateClient) TemplateInfoText(ctx context.Context, name string) (string, error) {
	return w.Run(ctx, "templates", "info", name, "--quiet")
}

// ValidateTemplate runs `warpgate validate`.
func (w *WarpgateClient) ValidateTemplate(ctx context.Context, config string, syntaxOnly bool) (string, error) {
	args := newArgs("validate", config).boolFlag("--syntax-only", syntaxOnly)
	return w.Run(ctx, args.build()...)
}

// BuildOptions covers the `warpgate build` flag surface.
type BuildOptions struct {
	// Source selectors (exactly one of Template, Config, FromGit must be set).
	Template string
	Config   string
	FromGit  string

	Target          string
	Architectures   []string
	Push            bool
	PushDigest      bool
	Registry        string
	Tags            []string
	Variables       map[string]string
	VarFiles        []string
	BuildArgs       map[string]string
	Labels          map[string]string
	NoCache         bool
	CacheFrom       []string
	CacheTo         []string
	OutputManifest  string
	SaveDigests     bool
	DigestDir       string
	DryRun          bool
	Force           bool
	Region          string
	Regions         []string
	CopyToRegions   []string
	ParallelRegions bool
	InstanceType    string
	StreamLogs      bool
	ShowEC2Status   bool
}

func (w *WarpgateClient) BuildTemplate(ctx context.Context, opts BuildOptions) (string, error) {
	args := newArgs("build")
	switch {
	case opts.Template != "":
		args.str("--template", opts.Template)
	case opts.FromGit != "":
		args.str("--from-git", opts.FromGit)
	case opts.Config != "":
		args.raw(opts.Config)
	default:
		return "", fmt.Errorf("build requires one of: template, config, or from_git")
	}

	args.
		str("--target", opts.Target).
		csv("--arch", opts.Architectures).
		boolFlag("--push", opts.Push).
		boolFlag("--push-digest", opts.PushDigest).
		str("--registry", opts.Registry).
		repeated("--tag", opts.Tags).
		kv("--var", opts.Variables).
		repeated("--var-file", opts.VarFiles).
		kv("--build-arg", opts.BuildArgs).
		kv("--label", opts.Labels).
		boolFlag("--no-cache", opts.NoCache).
		repeated("--cache-from", opts.CacheFrom).
		repeated("--cache-to", opts.CacheTo).
		str("--output-manifest", opts.OutputManifest).
		boolFlag("--save-digests", opts.SaveDigests).
		str("--digest-dir", opts.DigestDir).
		boolFlag("--dry-run", opts.DryRun).
		boolFlag("--force", opts.Force).
		str("--region", opts.Region).
		csv("--regions", opts.Regions).
		csv("--copy-to-regions", opts.CopyToRegions).
		boolFlag("--parallel-regions", opts.ParallelRegions).
		str("--instance-type", opts.InstanceType).
		boolFlag("--stream-logs", opts.StreamLogs).
		boolFlag("--show-ec2-status", opts.ShowEC2Status)

	return w.Run(ctx, args.build()...)
}

func (w *WarpgateClient) InitTemplate(ctx context.Context, name, fromTemplate, outputDir string) (string, error) {
	args := newArgs("init", name).
		str("--from", fromTemplate).
		str("--output", outputDir)
	return w.Run(ctx, args.build()...)
}

// AddTemplateSource registers a git URL or local path. For git URLs, name is optional
// (auto-derived if empty); for local paths name is ignored.
func (w *WarpgateClient) AddTemplateSource(ctx context.Context, urlOrPath, name string) (string, error) {
	args := newArgs("templates", "add")
	if name != "" {
		args.raw(name)
	}
	args.raw(urlOrPath)
	return w.Run(ctx, args.build()...)
}

func (w *WarpgateClient) RemoveTemplateSource(ctx context.Context, pathOrName string) (string, error) {
	return w.Run(ctx, "templates", "remove", pathOrName)
}

func (w *WarpgateClient) UpdateTemplateCache(ctx context.Context) (string, error) {
	return w.Run(ctx, "templates", "update")
}

// ConvertOptions covers `warpgate convert packer`.
type ConvertOptions struct {
	TemplateDir string
	Output      string
	Author      string
	Version     string
	BaseImage   string
	License     string
	IncludeAMI  *bool // nil = leave default
	DryRun      bool
}

func (w *WarpgateClient) ConvertPackerTemplate(ctx context.Context, opts ConvertOptions) (string, error) {
	args := newArgs("convert", "packer", opts.TemplateDir).
		str("--output", opts.Output).
		str("--author", opts.Author).
		str("--version", opts.Version).
		str("--base-image", opts.BaseImage).
		str("--license", opts.License).
		triBool("--include-ami", opts.IncludeAMI).
		boolFlag("--dry-run", opts.DryRun)
	return w.Run(ctx, args.build()...)
}

// ManifestOptions covers `warpgate manifests create`.
type ManifestOptions struct {
	ImageName             string
	Registry              string
	Namespace             string
	AuthFile              string
	Tags                  []string
	DigestDir             string
	RequiredArchitectures []string
	BestEffort            bool
	Annotations           map[string]string
	Labels                map[string]string
	DryRun                bool
	Force                 bool
	HealthCheck           bool
	MaxAge                string
	ShowDiff              bool
	VerifyRegistry        *bool // nil = leave default
	VerifyConcurrency     int
}

func (w *WarpgateClient) CreateManifest(ctx context.Context, opts ManifestOptions) (string, error) {
	args := newArgs("manifests").
		str("--registry", opts.Registry).
		str("--namespace", opts.Namespace).
		str("--auth-file", opts.AuthFile).
		raw("create").
		str("--name", opts.ImageName).
		repeated("--tag", opts.Tags).
		str("--digest-dir", opts.DigestDir).
		repeated("--require-arch", opts.RequiredArchitectures).
		boolFlag("--best-effort", opts.BestEffort).
		kv("--annotation", opts.Annotations).
		kv("--label", opts.Labels).
		boolFlag("--dry-run", opts.DryRun).
		boolFlag("--force", opts.Force).
		boolFlag("--health-check", opts.HealthCheck).
		str("--max-age", opts.MaxAge).
		boolFlag("--show-diff", opts.ShowDiff).
		triBool("--verify-registry", opts.VerifyRegistry).
		intFlag("--verify-concurrency", opts.VerifyConcurrency)
	return w.Run(ctx, args.build()...)
}

func (w *WarpgateClient) InspectManifest(ctx context.Context, name, registry, namespace string, tags []string) (string, error) {
	args := newArgs("manifests").
		str("--registry", registry).
		str("--namespace", namespace).
		raw("inspect").
		str("--name", name).
		repeated("--tag", tags)
	return w.Run(ctx, args.build()...)
}

func (w *WarpgateClient) ListManifests(ctx context.Context, name, registry, namespace string) (string, error) {
	args := newArgs("manifests").
		str("--registry", registry).
		str("--namespace", namespace).
		raw("list").
		str("--name", name)
	return w.Run(ctx, args.build()...)
}

// CleanupOptions covers `warpgate cleanup`.
type CleanupOptions struct {
	BuildName string
	All       bool
	DryRun    bool
	Region    string
	Versions  bool
	Keep      int
	Yes       bool
}

func (w *WarpgateClient) Cleanup(ctx context.Context, opts CleanupOptions) (string, error) {
	args := newArgs("cleanup")
	if opts.BuildName != "" {
		args.raw(opts.BuildName)
	}
	args.
		boolFlag("--all", opts.All).
		boolFlag("--dry-run", opts.DryRun).
		str("--region", opts.Region).
		boolFlag("--versions", opts.Versions)
	if opts.Versions {
		args.intFlag("--keep", opts.Keep)
	}
	args.boolFlag("--yes", opts.Yes)
	return w.Run(ctx, args.build()...)
}

func (w *WarpgateClient) ConfigShow(ctx context.Context) (string, error) {
	return w.Run(ctx, "config", "show")
}

func (w *WarpgateClient) ConfigPath(ctx context.Context) (string, error) {
	out, err := w.Run(ctx, "config", "path")
	return strings.TrimSpace(out), err
}

func (w *WarpgateClient) ConfigGet(ctx context.Context, key string) (string, error) {
	out, err := w.Run(ctx, "config", "get", key)
	return strings.TrimSpace(out), err
}

func (w *WarpgateClient) ConfigSet(ctx context.Context, key, value string) (string, error) {
	return w.Run(ctx, "config", "set", key, value)
}

func (w *WarpgateClient) Version(ctx context.Context) (string, error) {
	out, err := w.Run(ctx, "version")
	return strings.TrimSpace(out), err
}
