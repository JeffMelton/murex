# Murex Development Commands

## Build Commands
- `make build` - Build the Murex binary to ./bin/murex
- `make build-dev` - Build with debug symbols, race detection, and profiling support
- `make install` - Install Murex to /usr/bin (requires root)
- `make clean` - Remove build artifacts

## Development Tools
- `make run ARGS="..."` - Build and run development version with arguments
- `make generate` - Run code generation (`go generate ./...`)
- `make lint` - Run golangci-lint (requires golangci-lint installed)
- `make test` - Run Go tests and behavioral tests
- `make bench` - Run benchmarks

## Testing
- `go test ./... -count 1 -race -covermode=atomic` - Run Go unit tests
- `./bin/murex -c 'g behavioural/*.mx -> foreach f { source $$f }; test run *'` - Run behavioral tests
- Behavioral tests are written in Murex language (.mx files) in `/behavioural/` directory

## Dependencies
- `make update-deps` - Update Go dependencies
- `go mod tidy` - Clean up go.mod

## Documentation (VuePress)
- `npm run docs:build` - Build documentation website
- `npm run docs:dev` - Run development documentation server
- `npm run docs:clean-dev` - Clean cache and run dev server
- Uses `pnpm` as package manager

## Build Configuration
- `BUILD_TAGS` - Additional build tags (default from `builtins/optional/standard-opts.txt`)
- `GO_FLAGS` - Additional go build flags
- `make list-build-tags` - Show available build tags