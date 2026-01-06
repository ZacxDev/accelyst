# Accelyst Configuration Specification

Version: 1.0.0

## Overview

An Accelyst configuration file defines the structure of a project and the milestones required to evolve it. The file format is YAML.

## Schema

### Root Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectParts` | `ProjectPart[]` | Yes | Components of the project |
| `milestones` | `Milestone[]` | Yes | Ordered list of milestones to execute |

### ProjectPart

Defines a logical component of the project with its location and optional documentation.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Unique identifier for this project part |
| `directoryAbs` | `string` | Yes | Absolute path to the component's root directory |
| `documentationAbs` | `string` | No | Absolute path to documentation file or directory |

**Constraints:**
- `name` must be unique across all project parts
- `name` must be non-empty
- `directoryAbs` must be an absolute path

### Milestone

Defines a feature or goal composed of ordered steps.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Human-readable name for the milestone |
| `steps` | `Step[]` | Yes | Steps required to complete the milestone |

**Constraints:**
- `name` must be non-empty
- `steps` must contain at least one step

### Step

Defines a single action within a milestone.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `integer` | Yes | Unique identifier within the milestone |
| `instruction` | `string` | Yes | Action to perform |
| `dependsOnStepIds` | `integer[]` | No | Step IDs that must complete before this step |
| `projectParts` | `string[]` | No | Project part names relevant to this step |

**Constraints:**
- `id` must be unique within the milestone
- `id` must be a positive integer
- `dependsOnStepIds` must reference valid step IDs within the same milestone
- `dependsOnStepIds` must not create circular dependencies
- `projectParts` must reference valid project part names

## Execution Semantics

### Dependency Resolution

Steps form a directed acyclic graph (DAG) where edges represent dependencies. Accelyst performs topological sorting using Kahn's algorithm to determine execution order.

```
Step A (id: 1, dependsOnStepIds: [])     → executes first (no dependencies)
Step B (id: 2, dependsOnStepIds: [])     → executes first (no dependencies)
Step C (id: 3, dependsOnStepIds: [1, 2]) → executes after A and B complete
```

### Parallel Execution

Steps with no dependencies (`dependsOnStepIds` is empty or omitted) are candidates for parallel execution. These steps receive the `use a subagent to` prefix in generated prompts.

### Prompt Generation

For each step, a prompt fragment is generated with the following structure:

```
<id>: [subagent prefix] [project parts clause] [dependencies clause] <instruction>
```

**Components:**

1. **Subagent prefix** (if no dependencies):
   ```
   use a subagent to
   ```

2. **Project parts clause** (if projectParts specified):
   ```
   Analyze <name> (<directoryAbs>) [(docs: <documentationAbs>)], ... then
   ```

3. **Dependencies clause** (if dependsOnStepIds specified):
   ```
   read results from steps <id>, <id>, ... then
   ```

4. **Instruction**: The `instruction` field value

## Example

```yaml
projectParts:
  - name: api
    directoryAbs: /home/user/project/api
    documentationAbs: /home/user/project/docs/api.md
  - name: client
    directoryAbs: /home/user/project/client

milestones:
  - name: "Add User Authentication"
    steps:
      - id: 1
        instruction: "research JWT best practices"
        dependsOnStepIds: []
        projectParts: []
      - id: 2
        instruction: "analyze existing auth patterns"
        dependsOnStepIds: []
        projectParts:
          - api
      - id: 3
        instruction: "implement auth middleware"
        dependsOnStepIds:
          - 1
          - 2
        projectParts:
          - api
      - id: 4
        instruction: "implement login UI"
        dependsOnStepIds:
          - 3
        projectParts:
          - client
```

**Output:**

```
# Milestone: Add User Authentication

1: use a subagent to research JWT best practices
2: use a subagent to Analyze api (/home/user/project/api) (docs: /home/user/project/docs/api.md) then analyze existing auth patterns
3: Analyze api (/home/user/project/api) (docs: /home/user/project/docs/api.md) then read results from steps 1, 2 then implement auth middleware
4: Analyze client (/home/user/project/client) then read results from steps 3 then implement login UI
```

## JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["projectParts", "milestones"],
  "properties": {
    "projectParts": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "directoryAbs"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "directoryAbs": { "type": "string", "minLength": 1 },
          "documentationAbs": { "type": "string" }
        }
      }
    },
    "milestones": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "steps"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "steps": {
            "type": "array",
            "minItems": 1,
            "items": {
              "type": "object",
              "required": ["id", "instruction"],
              "properties": {
                "id": { "type": "integer", "minimum": 1 },
                "instruction": { "type": "string", "minLength": 1 },
                "dependsOnStepIds": {
                  "type": "array",
                  "items": { "type": "integer", "minimum": 1 }
                },
                "projectParts": {
                  "type": "array",
                  "items": { "type": "string" }
                }
              }
            }
          }
        }
      }
    }
  }
}
```
