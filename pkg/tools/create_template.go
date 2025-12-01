// Copyright (c) 2025 CowDogMoo
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cowdogmoo/warpgate-mcp-server/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func createTemplate(s *server.MCPServer, logger *logging.Logger, warpgatePath string) {
	tool := mcp.Tool{
		Name:        "create_template",
		Description: "Create a new Packer template with all required files and structure",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"template_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the new template (e.g., 'my-awesome-template')",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Brief description of what this template creates",
				},
				"base_image": map[string]interface{}{
					"type":        "string",
					"description": "Base Docker image to use (default: 'ubuntu')",
				},
				"base_image_version": map[string]interface{}{
					"type":        "string",
					"description": "Version of the base image (default: 'latest')",
				},
				"include_ami": map[string]interface{}{
					"type":        "boolean",
					"description": "Include AWS AMI configuration (default: false)",
				},
			},
			Required: []string{"template_name", "description"},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.Params.Arguments
		templateName := args["template_name"].(string)
		description := args["description"].(string)

		// Validate template name
		if !isValidTemplateName(templateName) {
			return mcp.NewToolResultError(
				"Invalid template name. Use lowercase letters, numbers, and hyphens only."), nil
		}

		// Set defaults
		baseImage := "ubuntu"
		if val, ok := args["base_image"].(string); ok && val != "" {
			baseImage = val
		}

		baseImageVersion := "latest"
		if val, ok := args["base_image_version"].(string); ok && val != "" {
			baseImageVersion = val
		}

		includeAMI := false
		if val, ok := args["include_ami"].(bool); ok {
			includeAMI = val
		}

		// Create template directory
		templateDir := filepath.Join(warpgatePath, "packer-templates", templateName)
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			logger.Errorf("Failed to create template directory: %v", err)
			return mcp.NewToolResultError(
				fmt.Sprintf("Failed to create template directory: %v", err)), nil
		}

		// Check if template already exists
		if _, err := os.Stat(filepath.Join(templateDir, "docker.pkr.hcl")); err == nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("Template '%s' already exists at %s", templateName, templateDir)), nil
		}

		// Create files
		files := map[string]string{
			"plugins.pkr.hcl":   generatePluginsFile(includeAMI),
			"locals.pkr.hcl":    generateLocalsFile(),
			"variables.pkr.hcl": generateVariablesFile(templateName, baseImage, baseImageVersion, includeAMI),
			"docker.pkr.hcl":    generateDockerFile(templateName, description),
			"README.md":         generateReadme(templateName, description, baseImage, includeAMI),
		}

		if includeAMI {
			files["ami.pkr.hcl"] = generateAMIFile(templateName, description)
		}

		for filename, content := range files {
			filePath := filepath.Join(templateDir, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				logger.Errorf("Failed to create %s: %v", filename, err)
				return mcp.NewToolResultError(
					fmt.Sprintf("Failed to create %s: %v", filename, err)), nil
			}
		}

		logger.Infof("Successfully created template: %s", templateName)
		result := fmt.Sprintf(`Successfully created template '%s' at %s

Files created:
- plugins.pkr.hcl (Packer plugin requirements)
- locals.pkr.hcl (Local variables)
- variables.pkr.hcl (Template variables)
- docker.pkr.hcl (Docker build configuration)
- README.md (Template documentation)
%s

Next steps:
1. Initialize the template: task template-init TEMPLATE_NAME=%s
2. Customize the provisioning steps in docker.pkr.hcl
3. Validate the template: task template-validate TEMPLATE_NAME=%s
4. Build the template: task template-build TEMPLATE_NAME=%s
`, templateName, templateDir,
	func() string {
		if includeAMI {
			return "- ami.pkr.hcl (AWS AMI configuration)\n"
		}
		return ""
	}(), templateName, templateName, templateName)

		return mcp.NewToolResultText(result), nil
	}

	s.AddTool(tool, handler)
}

func isValidTemplateName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func generatePluginsFile(includeAMI bool) string {
	plugins := `# Define the plugin(s) used by Packer.
packer {
  required_plugins {
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "~> 1"
    }
`
	if includeAMI {
		plugins += `    amazon = {
      source  = "github.com/hashicorp/amazon"
      version = "~> 1"
    }
    ansible = {
      source  = "github.com/hashicorp/ansible"
      version = "~> 1"
    }
`
	}
	plugins += `  }
}
`
	return plugins
}

func generateLocalsFile() string {
	return `locals {
  timestamp = formatdate("YYYY-MM-DD-hh-mm-ss", timestamp())
}
`
}

func generateVariablesFile(templateName, baseImage, baseImageVersion string, includeAMI bool) string {
	vars := fmt.Sprintf(`#######################################################
#                  Warpgate variables                 #
#######################################################
variable "template_name" {
  type        = string
  description = "Name of the packer template."
  default     = "%s"
}

variable "provision_repo_path" {
  type        = string
  description = "Path on disk to the repo that contains the provisioning code."
  default     = ""
}

variable "shell" {
  type        = string
  description = "Shell to use."
  default     = "/bin/bash"
}

############################################
#           Container variables            #
############################################
variable "base_image" {
  type        = string
  description = "Base image."
  default     = "%s"
}

variable "base_image_version" {
  type        = string
  description = "Version of the base image."
  default     = "%s"
}

variable "entrypoint" {
  type        = string
  description = "Optional entrypoint script."
  default     = ""
}

variable "manifest_path" {
  type        = string
  description = "Path to the generated manifest file."
  default     = "manifest.json"
}

variable "user" {
  type        = string
  description = "Default user."
  default     = "root"
}

variable "workdir" {
  type        = string
  description = "Working directory for a new container."
  default     = "/root"
}
`, templateName, baseImage, baseImageVersion)

	if includeAMI {
		vars += `
############################################
#              AWS variables               #
############################################
variable "ami_region" {
  type        = string
  description = "AWS region to launch the instance and create AMI."
  default     = "us-east-1"
}

variable "instance_type" {
  type        = string
  description = "The type of instance to use for the initial AMI creation."
  default     = "t3.micro"
}

variable "disk_size" {
  type        = number
  description = "Disk size in GB for building the AMI."
  default     = 50
}

variable "ssh_username" {
  type        = string
  description = "The SSH username for the AMI."
  default     = "ubuntu"
}

variable "ssh_timeout" {
  type        = string
  description = "Timeout for SSH connections."
  default     = "20m"
}
`
	}

	return vars
}

func generateDockerFile(templateName, description string) string {
	titleComment := fmt.Sprintf("#########################################################################################\n# %s packer template\n#\n# Description: %s\n#########################################################################################", templateName, description)

	return fmt.Sprintf(`%s
# Docker AMD64 source configuration
source "docker" "amd64" {
  commit     = true
  image      = "${var.base_image}:${var.base_image_version}"
  platform   = "linux/amd64"
  privileged = true

  volumes = {
    "/sys/fs/cgroup" = "/sys/fs/cgroup:rw"
  }

  changes = [
    "ENTRYPOINT ${var.entrypoint}",
    "USER ${var.user}",
    "WORKDIR ${var.workdir}",
  ]

  run_command = ["-d", "-i", "-t", "--cgroupns=host", "{{ .Image }}"]
}

# Docker ARM64 source configuration
source "docker" "arm64" {
  commit     = true
  image      = "${var.base_image}:${var.base_image_version}"
  platform   = "linux/arm64"
  privileged = true

  changes = [
    "ENTRYPOINT ${var.entrypoint}",
    "USER ${var.user}",
    "WORKDIR ${var.workdir}",
  ]

  volumes = {
    "/sys/fs/cgroup" = "/sys/fs/cgroup:rw"
  }

  run_command = ["-d", "-i", "-t", "--cgroupns=host", "{{ .Image }}"]
}

build {
  name = "%s-docker"
  sources = [
    "source.docker.amd64",
    "source.docker.arm64"
  ]

  # Add your provisioning steps here
  # Example: Install packages
  provisioner "shell" {
    only = ["docker.arm64", "docker.amd64"]
    inline = [
      "echo 'Add your provisioning commands here'",
      "# apt-get update && apt-get install -y <your-packages>",
    ]
  }

  # Clean up to reduce image size
  provisioner "shell" {
    only = ["docker.arm64", "docker.amd64"]
    inline = [
      "apt-get clean || true",
      "rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* || true"
    ]
  }

  # Create manifest with the necessary information to tag and push the created image(s)
  post-processor "manifest" {
    output     = "${var.manifest_path}"
    strip_path = true
  }
}
`, titleComment, templateName)
}

func generateAMIFile(templateName, description string) string {
	titleComment := fmt.Sprintf("#########################################################################################\n# %s AMI packer template\n#\n# Description: %s\n#########################################################################################", templateName, description)

	return fmt.Sprintf(`%s
# AWS AMI source configuration
source "amazon-ebs" "amd64" {
  ami_name      = "%s-{{timestamp}}"
  instance_type = var.instance_type
  region        = var.ami_region
  ssh_username  = var.ssh_username
  ssh_timeout   = var.ssh_timeout

  source_ami_filter {
    filters = {
      name                = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"] # Canonical
  }

  launch_block_device_mappings {
    device_name = "/dev/sda1"
    volume_size = var.disk_size
    volume_type = "gp3"
    delete_on_termination = true
  }

  tags = {
    Name = "%s"
    Built_By = "Packer"
    Built_At = "{{timestamp}}"
  }
}

build {
  name = "%s-ami"
  sources = [
    "source.amazon-ebs.amd64"
  ]

  # Add your provisioning steps here
  provisioner "shell" {
    inline = [
      "echo 'Add your AMI provisioning commands here'",
      "sudo apt-get update",
      "# sudo apt-get install -y <your-packages>",
    ]
  }

  # Clean up
  provisioner "shell" {
    inline = [
      "sudo apt-get clean",
      "sudo rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*"
    ]
  }
}
`, titleComment, templateName, templateName, templateName)
}

func generateReadme(templateName, description, baseImage string, includeAMI bool) string {
	title := strings.ReplaceAll(strings.Title(strings.ReplaceAll(templateName, "-", " ")), " ", "-")

	readme := fmt.Sprintf(`# Packer Build for %s

%s

---

## Requirements

- [Packer](https://www.packer.io/)
- Docker (for building Docker images)
`, title, description)

	if includeAMI {
		readme += `- AWS account & credentials (for AMI builds)
`
	}

	readme += `- Required Packer plugins:
  - ` + "`docker`\n"

	if includeAMI {
		readme += `  - ` + "`amazon`\n"
		readme += `  - ` + "`ansible`" + ` (optional, for complex provisioning)\n`
	}

	readme += `
---

## Variables

See ` + "`variables.pkr.hcl`" + ` for all configurable parameters.

Key variables:
- ` + "`template_name`" + `: Name of the template (default: ` + "`" + templateName + "`)\n" +
`- ` + "`base_image`" + `: Base Docker image (default: ` + "`" + baseImage + "`)\n" +
`- ` + "`base_image_version`" + `: Version of the base image (default: ` + "`latest`)\n"

	if includeAMI {
		readme += `- ` + "`ami_region`" + `: AWS region for AMI creation (default: ` + "`us-east-1`)\n"
	}

	readme += `
---

## Building Docker Images

Initialize the template:

` + "```bash" + `
export TASK_X_REMOTE_TASKFILES=1
task -y template-init TEMPLATE_NAME=` + templateName + `
` + "```" + `

Build the images:

` + "```bash" + `
export TASK_X_REMOTE_TASKFILES=1
task -y template-build TEMPLATE_NAME=` + templateName + ` ONLY='` + templateName + `-docker.docker.*'
` + "```" + `

---
`

	if includeAMI {
		readme += `
## Building AWS AMIs

` + "```bash" + `
export TASK_X_REMOTE_TASKFILES=1
task -y template-build TEMPLATE_NAME=` + templateName + ` ONLY='` + templateName + `-ami.amazon-ebs.*'
` + "```" + `

> 🛡️ Ensure your AWS credentials are configured.

---
`
	}

	readme += `
## Customization

1. Edit ` + "`docker.pkr.hcl`" + ` to add your provisioning steps
2. Modify ` + "`variables.pkr.hcl`" + ` to adjust defaults
3. Update this README with your specific requirements

---

## Notes

- Multi-arch Docker images (` + "`amd64`" + ` + ` + "`arm64`" + `) are built by default
- Customize provisioning in the ` + "`provisioner`" + ` blocks
- Images are suitable for CI, local testing, or deployment
`

	return readme
}
