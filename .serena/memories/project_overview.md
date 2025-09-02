# Murex Project Overview

## What is Murex
Murex is a smart shell (like bash/zsh/fish) written in Go that provides enhanced features and improved UX:

- **Type-aware pipelines**: Support for complex data formats like JSON, tables with intelligent processing
- **Usability improvements**: In-line spell checking, context-sensitive hints, auto-parsing man pages for completions
- **Better error handling**: Try/catch blocks, line numbers in errors, debugging frameworks built into the language
- **Smart data processing**: Enhanced tools that work intelligently with structured data without additional configuration

## Tech Stack
- **Primary Language**: Go 1.24.2
- **Module**: github.com/lmorg/murex
- **Key Dependencies**: 
  - readline/v4 for interactive shell
  - sqlite3/modernc.org/sqlite for data storage
  - fsnotify for file watching
  - Various data format libraries (YAML, TOML, HCL, JSON via mxj)
- **Documentation**: VuePress-based website with markdown generation from YAML templates
- **Platform Support**: Cross-platform (Linux, macOS, Windows, Plan9, BSD variants)

## Repository Structure
- `/builtins/` - Core shell builtins, types, pipes, optional features
- `/lang/` - Core language implementation (interpreter, processes, variables, types)
- `/shell/` - Interactive shell features (autocomplete, history, syntax highlighting)
- `/app/` - Application layer
- `/config/` - Configuration management
- `/debug/` - Debugging tools
- `/docs/` - Documentation source
- `/examples/` - Usage examples
- `/gen/` - Code generation
- `/integrations/` - Platform-specific integrations
- `/test/` - Test suite
- `/utils/` - Utility functions
- `/behavioural/` - Behavioral tests in Murex language (.mx files)