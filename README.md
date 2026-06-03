# bzltool

[![Build and Test](https://github.com/filmil/bzltool/actions/workflows/build-test.yml/badge.svg)](https://github.com/filmil/bzltool/actions/workflows/build-test.yml)

`bzltool` is a flexible, highly configurable CLI and TUI utility written in Go for orchestrating Bazel project bootstraps and feature scaffolding. It merges templated "fragments" from diverse git repositories into a unified project structure.

## Features

- **Dynamic Project Initialization**: Generates a unified Bazel project repository by seamlessly concatenating code fragments and dependencies fetched from configurable external template Git repositories.
- **Go Templates**: Full support for `text/template` parameter substitution (e.g., `{{.ProjectName}}`) across merged fragments.
- **TUI & CLI**: Built with `Cobra` for powerful CLI interaction and `Bubbletea` for intuitive, terminal-based fallback configuration prompts when configuration is missing.
- **XDG Standards**: Global repository configurations are managed elegantly via per-user `.config/` XDG standard directories.
- **Project Workspaces**: Local project states are continuously saved in `.config/bzltool/project_config.json`, empowering synchronized updates and AI tooling interaction.

## Usage

**Global Configuration**
Before running `bzltool`, create your global configuration detailing where templates should be fetched from:
```bash
mkdir -p ~/.config/bzltool
cat <<EOF > ~/.config/bzltool/config.json
{
  "template_repos": [
    {
      "url": "https://github.com/my-org/my-templates.git",
      "subdir": "templates"
    }
  ]
}
EOF
```

**Project Initialization**
Initialize a new project in the current directory. You can pass the project name via flag:
```bash
bzltool init --project_name="My Awesome Project"
```

Alternatively, you can provide a full JSON configuration (which sets languages, toolchains, and modules):
```bash
bzltool init --config=my_config.json
```

If neither is provided or data is missing, an interactive TUI will gracefully prompt you for the necessary information.

**Template Configuration (`template.json`)**
Template repositories can include a `template.json` at their root to control file merging behaviors:
```json
{
  "ignore": ["**/.DS_Store", "**/node_modules/**"],
  "raw": ["*.sh", "Makefile"],
  "conditions": {
    "python": ["**/python/**"],
    "go": ["**/go/**"]
  }
}
```
- `ignore`: Files matching these patterns will never be copied.
- `raw`: Files matching these patterns will be copied verbatim, bypassing the Go text templating engine (useful to avoid `{{ }}` conflicts in shell scripts).
- `conditions`: Files matching these patterns are conditionally included ONLY if the project configuration explicitly requests that language or toolchain.

## Building and Testing

This project leverages Bazel (`MODULE.bazel`) for all builds and tests to ensure reproducible environments. 

To run the unit and end-to-end integration tests:
```bash
bazel test //...
```

To build the executable:
```bash
bazel build //...
```
