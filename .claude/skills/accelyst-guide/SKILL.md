---
name: accelyst-guide
description: Execute and manage Accelyst CLI operations. Use when the user wants to run accelyst commands, initialize projects, run milestones, manage registry, create epics, or work with artifacts. Also use when explaining accelyst concepts or troubleshooting issues.
allowed-tools: Read, Write, Glob, Grep, Bash
---

# Accelyst Operations Guide

Execute Accelyst CLI commands and manage project configuration for AI-native project iteration.

## Operational Instructions

When this skill is activated, you should:

1. **Locate the accelyst binary** - Check for `./accelyst` in project root or use the built binary path
2. **Check project state** - Verify `.accelyst/` directory exists before running commands
3. **Execute commands** - Use Bash tool to run accelyst CLI commands
4. **Read/write config** - Use Read/Write tools to inspect and modify YAML files
5. **Report results** - Show command output and explain what happened

### Binary Location

The accelyst binary is typically at the project root. Before running commands:

```bash
# Check if binary exists and is executable
ls -la ./accelyst

# If not built, build it first
go build -o accelyst
```

## Action Patterns

### Initialize a New Project

```bash
# 1. Build the binary if needed
go build -o accelyst

# 2. Initialize .accelyst/ structure
./accelyst init

# 3. Verify initialization
ls -la .accelyst/
```

Then help the user edit `.accelyst/registry.yaml` to add their project parts.

### List Available Epics

```bash
./accelyst epics list
```

### Run a Milestone

```bash
# Run entire milestone - outputs execution prompt
./accelyst run <epic-name>/<milestone-id>

# Example:
./accelyst run auth-feature/jwt_auth
```

The output is an execution-ready prompt. Present it to the user or execute it.

### Run a Specific Step

```bash
./accelyst run <epic-name>/<milestone-id>/<step-id>

# Example:
./accelyst run auth-feature/jwt_auth/analyze_auth
```

### Inspect Registry

```bash
# List project parts
./accelyst registry parts

# List artifacts
./accelyst registry artifacts
```

### Add an Artifact to Registry

```bash
./accelyst registry add-artifact <id> "<description>" <epic/milestone/step>

# Example:
./accelyst registry add-artifact auth_analysis "Authentication flow analysis" auth-feature/jwt_auth/analyze_auth
```

### Migrate from Legacy Format

```bash
./accelyst migrate
```

## Configuration Management

### Reading Configuration

```bash
# Check registry
cat .accelyst/registry.yaml

# List epics
ls .accelyst/epics/

# Read specific epic
cat .accelyst/epics/<name>.yaml

# Check artifacts directory
ls .accelyst/artifacts/
```

### Creating/Editing Registry

Use Write tool to create or update `.accelyst/registry.yaml`:

```yaml
version: 1

projectParts:
  - name: client
    directoryAbs: /absolute/path/to/client
    description: "Frontend application"
  - name: api
    directoryAbs: /absolute/path/to/api
    documentationAbs: /absolute/path/to/docs/api.md
    description: "Backend API"
  - name: test
    directoryAbs: /absolute/path/to/tests
    description: "Test suites"

artifacts: []
```

**Important**: `directoryAbs` must be absolute paths. Use `pwd` to get current directory if needed.

### Creating Epic Files

Use Write tool to create `.accelyst/epics/<epic-name>.yaml`:

```yaml
name: "Feature Name"
description: "What this epic accomplishes"
priority: ready  # ready | in_progress | blocked | ideas

milestones:
  - id: milestone_id
    name: "Milestone Name"
    dependsOn: []
    steps:
      - id: step_id
        instruction: "action to perform"
        projectParts:
          - client
```

### Creating Artifacts

After running a step with `produces`, save artifact to `.accelyst/artifacts/<id>.md`:

```markdown
# Artifact: <artifact_id>

## Summary
[Key findings from analysis]

## Component Locations
- File: path/to/file.ts
- Function: functionName

## Key Interfaces
[Important types and contracts]

## Extension Points
[Where to add new functionality]
```

## Workflow Execution

### Complete Workflow: New Feature

1. **Check current state**:
   ```bash
   ./accelyst epics list
   ./accelyst registry parts
   ```

2. **Create epic file** (use Write tool):
   - Path: `.accelyst/epics/<feature-name>.yaml`
   - Include milestones with steps

3. **Validate by listing**:
   ```bash
   ./accelyst epics list
   ```

4. **Run first milestone**:
   ```bash
   ./accelyst run <epic>/<first-milestone>
   ```

5. **Execute the generated prompt** or present to user

6. **If step produces artifact**, save it to `.accelyst/artifacts/`

7. **Continue with next milestone/step**

### Complete Workflow: Execute Milestone Prompts

When running a milestone, accelyst outputs structured prompts. For each step:

1. **Run the step**:
   ```bash
   ./accelyst run epic/milestone/step
   ```

2. **Parse the output** - Contains:
   - Project part paths to analyze
   - Dependencies to read results from
   - Instruction to execute
   - Artifact save location (if `produces`)

3. **Execute the instruction** using appropriate tools

4. **Save artifacts** if the step has `produces`

5. **Continue to next step**

## Architecture Reference

### Split Configuration Model

```
project/
├── .accelyst/
│   ├── registry.yaml      # Persistent: project parts + artifacts
│   ├── epics/             # Ephemeral: milestone collections
│   │   ├── feature-a.yaml
│   │   └── feature-b.yaml
│   └── artifacts/         # Knowledge outputs
│       └── analysis.md
```

| Component | Purpose | Lifecycle |
|-----------|---------|-----------|
| **Registry** | Project parts, registered artifacts | Persistent |
| **Epics** | Milestone collections | Ephemeral |
| **Artifacts** | Analysis outputs | Persistent |

### Step Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique within milestone (snake_case) |
| `instruction` | Yes | Action to perform |
| `dependsOn` | No | Step IDs that must complete first |
| `projectParts` | No | Components relevant to this step |
| `produces` | No | Artifact ID this step creates |
| `requires` | No | Artifact IDs needed as context |

### Output Format

Accelyst generates prompts with these components:

- `use a subagent to` - Prefix for parallelizable steps (no dependencies)
- `Analyze <name> (<path>)` - Project part context
- `### Context: <artifact>` - Injected artifact content
- `read results from steps X, Y` - Dependency chain
- `Save analysis to: .accelyst/artifacts/<id>.md` - Artifact save instruction

## Error Handling

### "No .accelyst directory"

```bash
./accelyst init
```

### "Project part not found"

Check registry and fix the name:
```bash
./accelyst registry parts
cat .accelyst/registry.yaml
```

### "Artifact not found"

Either add to registry or ensure a prior step produces it:
```bash
./accelyst registry artifacts
./accelyst registry add-artifact <id> "<desc>" <path>
```

### "Circular dependency detected"

Review and fix the dependency graph in the epic YAML file.

### Build errors

```bash
go build -o accelyst
```

## Quick Command Reference

| Operation | Command |
|-----------|---------|
| Build binary | `go build -o accelyst` |
| Initialize | `./accelyst init` |
| List epics | `./accelyst epics list` |
| Run milestone | `./accelyst run epic/milestone` |
| Run step | `./accelyst run epic/milestone/step` |
| List parts | `./accelyst registry parts` |
| List artifacts | `./accelyst registry artifacts` |
| Add artifact | `./accelyst registry add-artifact id desc path` |
| Migrate | `./accelyst migrate` |

## Related Skills

- `/milestone-orchestrate` - **Full-cycle**: backlog → implemented features (orchestrates all skills below)
- `/milestone-plan` - Interactive backlog discovery → planning brief
- `/milestone-convert` - Convert backlogs → epic YAML
- `/milestone-execute` - Orchestrate milestone execution via subagents
