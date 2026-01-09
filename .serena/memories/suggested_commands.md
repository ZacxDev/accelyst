# Suggested Commands

## Build & Run
```bash
# Build the binary
go build -o accelyst

# Run directly
go run main.go <command>
```

## Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestName ./...
```

## CLI Commands
```bash
# Initialize a new project
./accelyst init

# List all epics
./accelyst epics list

# Run a milestone
./accelyst run <epic>/<milestone>

# Run a specific step
./accelyst run <epic>/<milestone>/<step>

# Registry management
./accelyst registry parts         # List project parts
./accelyst registry artifacts     # List registered artifacts
./accelyst registry add-artifact  # Register an artifact

# Migration from legacy format
./accelyst migrate
```

## Development Utilities
```bash
# Format code
go fmt ./...

# Lint (if golint installed)
golint ./...

# Static analysis
go vet ./...

# Dependencies
go mod tidy
```
