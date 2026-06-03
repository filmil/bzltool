# Gemini Project Instructions

When working on this project, please adhere to the following constraints and
guidelines as defined in the `ai/spec.md` specification:

## Core Technologies
* **Build System:** Use Bazel (`MODULE.bazel`) for all builds and tests.
* **Language:** Use Go.
* **Templating:** Use Go `text/template` for generating configuration files.
* **CLI & TUI:** Use the Cobra library for flag/config management, and
  Bubbletea for TUI elements when used interactively.

## Code Standards & Testing
* **Documentation:** All public APIs must be documented according to Go
  documentation standards (godoc comments).
* **Testing:** Every feature must have at least one unit test confirming its
  functionality. Aim for 100% code path coverage. Every unit test shall include
  multiple test cases that cover regular paths, error paths, and edge cases.
* **Commit Messages:** Use Conventional Commits v1.0.0 for all git commit
  messages (e.g., `feat:`, `fix:`, `docs:`, `test:`).

## Feature Details
* The tool operates on template repositories containing `skills/` and
  `fragments/` directories.
* Fragments are concatenated alphanumerically by path if defined across
  multiple sources.
* `bzltool init` accepts `--project_name` and `--config` (JSON format matching
  the spec preamble). The `project_name` should be made available to Go
  templates via `{{.ProjectName}}`.

Always prioritize updating and consulting `ai/spec.md` when new requirements
arise.
