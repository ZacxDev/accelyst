# Accelyst

An AI-native project iteration interface.

## Overview

Accelyst transforms your project structure and feature backlog into execution-ready prompts with full codebase context. Each prompt includes relevant directories, documentation paths, dependency ordering, and artifact context—everything an AI agent needs to implement features autonomously.

- **Rich context**: Prompts reference exact directories and documentation
- **Dependency-aware**: Milestones and steps are topologically sorted via DAG
- **Parallel-ready**: Independent work is marked for concurrent subagent execution
- **Knowledge persistence**: Analysis artifacts accumulate across iterations
- **Multiple workstreams**: Support for concurrent epics

## Installation

```bash
go build -o accelyst
```

## Quick Start

```bash
# Initialize a new project
./accelyst init

# List all epics
./accelyst epics list

# Run a milestone
./accelyst run <epic>/<milestone>

# Run a specific step
./accelyst run <epic>/<milestone>/<step>
```

## Architecture

Accelyst uses a split configuration model:

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

- **Registry** (`.accelyst/registry.yaml`): Persistent project knowledge including components and analysis artifacts
- **Epics** (`.accelyst/epics/*.yaml`): Disposable milestone collections that can be regenerated
- **Artifacts** (`.accelyst/artifacts/*.md`): Analysis outputs that persist across iterations

## Registry Format

```yaml
version: 1

projectParts:
  - name: client
    directoryAbs: /path/to/client
    description: "Frontend application"
  - name: api
    directoryAbs: /path/to/api
    documentationAbs: /path/to/docs/api.md
    description: "Backend API"
  - name: test
    directoryAbs: /path/to/tests
    description: "Test suites"

artifacts:
  - id: auth_flow
    description: "Authentication flow analysis"
    producedBy: auth-feature/jwt_auth/analyze_auth
    createdAt: "2024-01-08"
```

## Epic Format

```yaml
name: "Feature Name"
description: "What this epic accomplishes"
priority: ready  # ready | in_progress | blocked | ideas

milestones:
  - id: auth_api
    name: "Authentication API"
    dependsOn: []
    steps:
      - id: analyze_auth
        instruction: "analyze existing authentication patterns"
        produces: auth_analysis  # Creates artifact
        projectParts:
          - api

      - id: implement
        instruction: "implement auth endpoints"
        requires:
          - auth_analysis  # Uses artifact
        dependsOn:
          - analyze_auth
        projectParts:
          - api

      - id: test_auth
        instruction: "write authentication tests"
        dependsOn:
          - implement
        projectParts:
          - test

  - id: auth_ui
    name: "Authentication UI"
    dependsOn:
      - auth_api
    steps:
      - id: build_form
        instruction: "implement login form"
        requires:
          - auth_analysis  # Reuse artifact from previous milestone
        projectParts:
          - client
```

### Step Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier within milestone |
| `instruction` | Yes | Action to perform |
| `dependsOn` | No | Step IDs that must complete first |
| `projectParts` | No | Project parts relevant to this step |
| `produces` | No | Artifact ID this step creates |
| `requires` | No | Artifact IDs this step needs as context |

## Artifact System

Steps can produce and consume artifacts for knowledge transfer:

**Production**: A step with `produces: auth_analysis` will:
1. Include a save instruction in the prompt
2. Agent writes to `.accelyst/artifacts/auth_analysis.md`
3. Artifact is registered for future use

**Consumption**: A step with `requires: [auth_analysis]` will:
1. Load artifact content from disk
2. Inject it into the prompt context
3. Agent receives prior analysis as background

This enables knowledge accumulation across milestones and iterations.

## CLI Reference

```bash
# Project initialization
accelyst init              # Create .accelyst/ structure
accelyst migrate           # Migrate from legacy accelyst.yaml

# Epic management
accelyst epics list        # List all epics and milestones

# Execution
accelyst run <epic>/<milestone>         # Generate milestone prompt
accelyst run <epic>/<milestone>/<step>  # Generate single step prompt

# Registry inspection
accelyst registry parts         # List project parts
accelyst registry artifacts     # List registered artifacts
accelyst registry add-artifact  # Register an artifact
```

## Output Format

```
# Milestone: Authentication API

## Step: analyze_auth

use a subagent to Analyze api (/path/to/api) then analyze existing authentication patterns

---
Save analysis to: .accelyst/artifacts/auth_analysis.md

---

## Step: implement

### Context: auth_analysis

[contents of auth_analysis.md artifact]

---

Analyze api (/path/to/api) then read results from steps analyze_auth then implement auth endpoints
```

**Output components:**
- `use a subagent to` prefix marks parallelizable steps (no dependencies)
- `Analyze <name> (<path>)` provides codebase context
- `### Context:` sections inject required artifacts
- `read results from steps X, Y` chains dependent steps
- `Save analysis to:` instructs artifact creation

## Migration

Projects using the legacy single-file format can migrate:

```bash
./accelyst migrate
```

This will:
1. Create `.accelyst/` directory structure
2. Extract `projectParts` to `.accelyst/registry.yaml`
3. Move milestones to `.accelyst/epics/default.yaml`
4. Backup original file to `accelyst.yaml.bak`

## How It Works

1. Load registry and scan epics directory
2. Validate all references (project parts, artifacts, dependencies)
3. Topologically sort milestones by dependencies (Kahn's algorithm)
4. For each milestone:
   - Build a DAG from step dependencies
   - Topologically sort steps
   - Inject artifact content for `requires`
   - Generate prompt fragments with resolved references
   - Add save instructions for `produces`
5. Output concatenated prompts per milestone

## Claude Code Skills

Accelyst includes Claude Code skills for automated milestone management:

| Skill | Purpose |
|-------|---------|
| `/milestone-orchestrate` | Full-cycle: backlog → implemented features |
| `/milestone-plan` | Interactive backlog discovery → planning brief |
| `/milestone-convert` | Convert planning brief → epic YAML |
| `/milestone-execute` | Execute epic via subagents |
| `/accelyst-guide` | CLI operations and troubleshooting |

```bash
# Full automation from scratch file
/milestone-orchestrate scratch.txt

# Or run individual steps
/milestone-plan scratch.txt
/milestone-convert planning-brief.yaml
/milestone-execute my-epic/my-milestone
```

## Documentation

- [SPEC.md](SPEC.md) - Complete specification with JSON schemas
- [.claude/skills/](.claude/skills/) - Claude Code skill definitions
