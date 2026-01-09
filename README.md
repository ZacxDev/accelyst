# Accelyst

An AI-native project iteration interface.

## Overview

Accelyst transforms your project structure and feature backlog into execution-ready prompts with full codebase context. Each prompt includes relevant directories, documentation paths, dependency ordering, and completion criteria—everything an AI agent needs to implement features autonomously.

- **Rich context**: Prompts reference exact directories and documentation
- **Dependency-aware**: Milestones and steps are topologically sorted via DAG
- **Parallel-ready**: Independent work is marked for concurrent subagent execution

## Installation

```bash
go build -o accelyst
```

## Usage

```bash
./accelyst                      # uses accelyst.yaml
./accelyst -config custom.yaml  # custom config file
```

## Configuration Format

```yaml
projectParts:
  - name: client
    directoryAbs: /path/to/client
    documentationAbs: /path/to/docs/client.md  # optional
  - name: server
    directoryAbs: /path/to/server

milestones:
  - id: auth_api
    name: "Authentication API"
    dependsOn: []
    steps:
      - id: research
        instruction: "research JWT best practices"
        dependsOn: []
        projectParts: []
      - id: implement
        instruction: "implement auth endpoints"
        dependsOn:
          - research
        projectParts:
          - server

  - id: auth_ui
    name: "Authentication UI"
    dependsOn:
      - auth_api
    steps:
      - id: build_form
        instruction: "implement login form"
        dependsOn: []
        projectParts:
          - client
      - id: configure_oauth
        instruction: "configure OAuth provider in production"
        dependsOn:
          - build_form
```

### Fields

**Milestone:**
| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier across all milestones |
| `name` | Yes | Human-readable name |
| `dependsOn` | No | Milestone IDs that must complete first |
| `steps` | Yes | Ordered list of steps |

**Step:**
| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier within the milestone |
| `instruction` | Yes | Action to perform |
| `dependsOn` | No | Step IDs that must complete first |
| `projectParts` | No | Project parts relevant to this step |

## Output Format

```
# Milestone: Authentication API

research: use a subagent to research JWT best practices [done when: documented token expiry strategy]
implement: Analyze server (/path/to/server) then read results from steps research then implement auth endpoints [done when: login and logout endpoints working; unit tests passing]

---

# Milestone: Authentication UI (depends on: auth_api)

build_form: use a subagent to Analyze client (/path/to/client) then implement login form
configure_oauth: read results from steps build_form then configure OAuth provider in production
```

**Output components:**
- `use a subagent to` prefix marks parallelizable steps (no dependencies)
- `Analyze <name> (<path>)` provides codebase context
- `read results from steps X, Y` chains dependent steps

## How It Works

1. Parse `projectParts` and `milestones` from YAML
2. Topologically sort milestones by dependencies (Kahn's algorithm)
3. For each milestone:
   - Build a DAG from step dependencies
   - Topologically sort steps
   - Generate prompt fragments with resolved references
4. Output concatenated prompts per milestone
