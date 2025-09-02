# Murex Code Style and Conventions

## Go Code Style
- Standard Go formatting (gofmt)
- Uses golangci-lint for linting (though no custom config file found)
- Race detection enabled in development builds
- Code generation used extensively (`go generate ./...`)

## Project Conventions

### File Naming
- Platform-specific files: `*_platform.go` (unix, windows, plan9, js)
- Test files: `*_test.go`
- Fuzz test files: `*_fuzz_test.go` 
- Integration files: `xxx_platform.mx` where platform is `any`, `posix`, `linux`, `darwin`, etc.

### Package Structure
- Core language features in `/lang/`
- Shell-specific features in `/shell/`  
- Built-in commands in `/builtins/`
- Platform integrations in `/integrations/`

### Documentation Generation
- All documentation generated from `*_doc.yaml` template files
- Run `go generate ./...` to regenerate markdown from YAML templates
- Each generated markdown has footer linking to source YAML

### Testing Approach
- Go unit tests with race detection
- Behavioral tests written in Murex language (.mx files)
- Benchmark tests
- Coverage reporting with atomic mode

### Build Tags
- Optional features controlled via build tags
- Standard options in `builtins/optional/standard-opts.txt`
- Development builds include: "pprof,trace,no_crash_handler"

### Branching
- Pull requests should target `develop` branch
- `master` is main/stable branch