# Code Style & Conventions

## Go Conventions
- Standard Go formatting (`go fmt`)
- CamelCase for exported identifiers
- camelCase for unexported identifiers
- YAML struct tags use lowercase snake_case: `yaml:"fieldName"`

## Project Structure
- Single `main.go` with all implementation
- `main_test.go` for tests
- JSON schemas embedded as constants (`registrySchema`, `epicSchema`)

## Key Types
```go
// Core data structures
type ProjectPart struct { Name, DirectoryAbs, DocumentationAbs, Description string }
type Artifact struct { ID, Description, ProducedBy, CreatedAt string }
type Registry struct { Version int; ProjectParts []ProjectPart; Artifacts []Artifact }
type Step struct { ID, Instruction string; DependsOn, ProjectParts []string; Produces string; Requires []string }
type Milestone struct { ID, Name string; DependsOn []string; Steps []Step }
type Epic struct { Name, Description, Priority string; Milestones []Milestone; filename string }
type LegacyConfig struct { ProjectParts []ProjectPart; Milestones []Milestone }
```

## Key Functions
- `loadRegistry()`, `saveRegistry()` - Registry persistence
- `loadEpics()`, `loadEpic()` - Epic loading
- `topologicalSort()`, `topologicalSortMilestones()` - DAG sorting (Kahn's algorithm)
- `generatePromptFragment()` - Prompt generation
- `processMilestone()` - Milestone processing
- `cmd*()` functions - CLI command handlers

## Naming Patterns
- `load*` - Functions that read from disk
- `save*` - Functions that write to disk
- `validate*` - Validation functions
- `generate*` - Output generation
- `cmd*` - CLI command handlers
