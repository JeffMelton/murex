# Task Completion Workflow

## When a coding task is completed, run these commands in order:

### 1. Code Generation (if applicable)
```bash
make generate
```
- Regenerates code from templates
- Updates documentation from YAML sources
- Must be run if any `*_doc.yaml` files were modified

### 2. Linting
```bash
make lint
```
- Runs golangci-lint
- Requires golangci-lint to be installed
- Checks code style and common issues

### 3. Testing
```bash
make test
```
- Runs Go unit tests with race detection
- Runs behavioral tests written in Murex language
- Must pass before submitting changes

### 4. Build Verification
```bash
make build
```
- Ensures the project builds successfully
- Creates binary in `./bin/murex`

## Optional Additional Checks

### Benchmarks (for performance-related changes)
```bash
make bench
```

### Development Build (for testing with debug features)
```bash
make build-dev
make run ARGS="test-commands-here"
```

## Notes
- All tests must pass before submitting PR
- PRs should target `develop` branch
- Behavioral tests (.mx files) are written in Murex language
- Documentation changes require running `make generate` to update markdown files