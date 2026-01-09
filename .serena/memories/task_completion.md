# Task Completion Checklist

## Before Marking Complete
1. **Format code**: `go fmt ./...`
2. **Run vet**: `go vet ./...`
3. **Run tests**: `go test -v ./...`
4. **Verify build**: `go build -o accelyst`

## Quality Checks
- All tests pass
- No vet warnings
- Code is formatted
- Documentation updated if needed (README.md, SPEC.md)

## Testing Considerations
- Test file: `main_test.go`
- Tests use temp directories for isolation
- Test both valid and invalid configurations
- Test DAG sorting for cycle detection
- Test artifact production and consumption
