# bzltool: a bazel build template generator

This is an automated, either interactive, or pre-generated config based
bazel project initializer.

For example, it should be possible to tell the tool to generate a bazel project
for a python project, with a pybind11 C++ extension, and a rust library. Maybe
to install a toolchain as well.

The tool should be set up as a swiss-army knive command tool, such as

```
bzltool init --option1=... ...
```
```
bzltool add module --lang=python --name=server
bzltool add module --lang=python --name=client
bzltool add module --lang=cpp --name=pybind_extension --type=pybind11
bzltool add module --lang=rust --name=lib
bzltool add toolchain --lang=go --version=1.21
```

Also allow the options to be supplied as a JSON structure in a flag, so it
can be used by AI tooling. If an option class or a flag is not provided, then
the tool should prompt the user for the missing information with an interactive
TUI.

```
{
  "init": {
    "project_name": "my_project",
    "languages": ["python", "cpp", "rust"],
    "toolchains": [
      {"lang": "go", "version": "1.21"}
    ],
    "modules": [
      {"lang": "python", "name": "server"},
      {"lang": "python", "name": "client"},
      {"lang": "cpp", "name": "pybind_extension", "type": "pybind11"},
      {"lang": "rust", "name": "lib"}
    ]
  }
}
```

The makes a directory .config/bzltool in the project, where the current project
configuration is stored. When the user starts the tool, they modify the configuration.

The tool should be allowed to install AI skills as well on demand.

`bzltool commit` syncs the project up with the .config/bzltool directory.

## C.1 Constraints

* Bazel as build system.
* Go as implementation language.
* Go text templating as a way to generate files where necessary.
* Public APIs documented based on go documentation standards
* Git for version management.
* Conventional commits v1.0.0 for git commit messages.
* Bubbletea for TUI elements when used interactively.
* Cobra library for flag and config management.
* Every feature must have at least one unit test confirming its functionality,
  aim for 100% code path coverage. Every
  unit test shall have multiple test cases exercising both regular paths,
  and error paths, and edge cases.


## Requirements

Requirements are built out iteratively, and are therefore enumeraed.

### R.1 Basics

* The user can provide a source git repo for templates. This repo has specific
  directory structure:

  ```
  - 01.dir1/
    - skills/
    - fragments/
  ```

  For now this is all, but we may add more later

  The`skills` directory contains AI skills that can be installed.
  The `fragments` directory contains file fragments used to generate files.

  For example, if there is a file `01.dir1/fragments/GEMINI.md` that fragment
  shall be added to the GEMINI file in the project root directory.

  If multiple fragments dirs define the samefile, then the tool should merge
  them. The merge strategy is to concatenate them in the alphanumeric order of
  their respective paths. This means the contents of
  `01.dir/fragments/foo.txt` will come before `02.dir/fragments/foo.txt`.

  Similarly `01.dir/fragments/dir1/bar.txt` shall be used to generate
  `//dir1/bar.txt` in the project repository.

* The fragment files shall be processed as text templates, and it shall be
  allowed for them to have access to project parameters as a struct for
  substitution.

* The program shall have a JSON config file stored based on XDG standards.

* The config shall name a list of git repository sources which will be used
  as template repos. Multiple can be allowed, and the program shall shell out
  to git to check them out into the config dir.

* For language setups, we will for the time being use `fragments` dir.

### R.2 Project Name Injection

* The `init` command shall accept a `--project_name` string flag.
* Alternatively, the user can supply the project configuration as a JSON structure via a flag (e.g., `--config`), which includes `project_name` under the `init` key, matching the preamble example.
* The tool shall make the `project_name` parameter available to the text templates in the `params` struct (e.g. as `{{.ProjectName}}`) during fragment generation.
* Update existing code to pass this project name into the template execution.
