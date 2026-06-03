# bzltool

[![Build and Test](https://github.com/filmil/bzltool/actions/workflows/build-test.yml/badge.svg)](https://github.com/filmil/bzltool/actions/workflows/build-test.yml)

`bzltool` is a flexible, highly configurable CLI and TUI utility written in Go for orchestrating Bazel project bootstraps and feature scaffolding. It merges templated "fragments" from diverse git repositories into a unified project structure.

## Features

- **Dynamic Project Initialization**: Generates a unified Bazel project repository by seamlessly concatenating code fragments and dependencies fetched from configurable external template Git repositories.
- **Go Templates**: Full support for `text/template` parameter substitution (e.g., `{{.ProjectName}}`) across merged fragments.
- **TUI & CLI**: Built with `Cobra` for powerful CLI interaction and `Bubbletea` for intuitive, terminal-based fallback configuration prompts when configuration is missing.
- **XDG Standards**: Global repository configurations are managed elegantly via per-user `.config/` XDG standard directories.
- **Project Workspaces**: Local project states are continuously saved in `.config/bzltool/project_config.json`, empowering synchronized updates and AI tooling interaction.

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
