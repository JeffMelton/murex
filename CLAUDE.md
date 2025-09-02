# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Murex is a smart shell written in Go that enhances traditional shell capabilities with type-aware pipelines, improved error handling, and intelligent data processing. It supports complex data formats like JSON and tables while maintaining compatibility with existing UNIX tools.

## Development Commands

### Essential Build Commands
- `make build` - Build Murex binary to ./bin/murex
- `make build-dev` - Build with debug symbols, race detection, and profiling
- `make test` - Run both Go unit tests and Murex behavioral tests
- `make lint` - Run golangci-lint (requires golangci-lint installed)
- `make generate` - Run code generation (required after modifying *_doc.yaml files)

### Development Workflow
- `make run ARGS="..."` - Build and run development version
- `make clean` - Remove build artifacts
- `make bench` - Run benchmarks

### Documentation
- `npm run docs:dev` - Run VuePress documentation development server
- `npm run docs:build` - Build documentation website
- Uses `pnpm` as package manager

## Architecture

### Core Directories
- **`/lang/`** - Language implementation (interpreter, processes, variables, types)
- **`/shell/`** - Interactive shell features (autocomplete, history, syntax highlighting)  
- **`/builtins/`** - Built-in shell commands and data type implementations
- **`/behavioural/`** - Behavioral tests written in Murex language (.mx files)
- **`/integrations/`** - Platform-specific integration files (format: `name_platform.mx`)

### Key Components
- **Process Management**: Core execution engine in `/lang/process.go`
- **Type System**: Sophisticated type handling for pipelines in `/lang/types/`
- **Interactive Features**: Autocomplete, hints, and syntax highlighting in `/shell/`
- **Built-ins**: Extensible command system in `/builtins/core/`

## Code Conventions

### Go Standards
- Standard Go formatting and conventions
- Race detection enabled in development builds
- Code generation extensively used (`go generate ./...`)

### File Patterns
- Platform-specific: `*_platform.go` (unix, windows, plan9, js)
- Tests: `*_test.go`, `*_fuzz_test.go`
- Documentation: Generated from `*_doc.yaml` templates

### Testing Strategy
- Go unit tests with race detection and atomic coverage mode
- Behavioral tests in Murex language (.mx files) 
- Command: `./bin/murex -c 'g behavioural/*.mx -> foreach f { source $$f }; test run *'`

## Task Completion Checklist

After completing any code changes, run:
1. `make generate` (if documentation or templates were modified)
2. `make lint` 
3. `make test`
4. `make build`

## Build Configuration

### Build Tags
- Standard options defined in `builtins/optional/standard-opts.txt`
- Development builds include: "pprof,trace,no_crash_handler"
- Use `make list-build-tags` to see available options

### Variables
- `BUILD_TAGS` - Additional build tags
- `GO_FLAGS` - Additional go build flags

## Contributing Notes

- Pull requests should target `develop` branch
- Documentation is generated from YAML templates - modify `*_doc.yaml` files, not markdown directly
- Integration files follow pattern: `name_platform.mx` where platform is `any`, `posix`, `linux`, `darwin`, etc.
- Project maintains excellent backwards compatibility commitment