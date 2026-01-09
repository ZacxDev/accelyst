# Accelyst Configuration Specification

Version: 2.0.0

## Overview

Accelyst uses a split configuration model:
- **Registry** (`.accelyst/registry.yaml`): Persistent project knowledge (parts, artifacts)
- **Epics** (`.accelyst/epics/*.yaml`): Ephemeral milestone groups

This separation enables knowledge accumulation across project iterations while keeping work items disposable.

## File Structure

```
project/
├── .accelyst/
│   ├── registry.yaml           # Persistent project knowledge
│   ├── epics/                   # Ephemeral work items
│   │   ├── feature-a.yaml
│   │   └── feature-b.yaml
│   └── artifacts/               # Analysis outputs
│       ├── component_structure.md
│       └── auth_flow.md
```

---

## Registry Schema

The registry is the persistent source of truth for project-level definitions.

### Root Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | `integer` | Yes | Schema version (currently `1`) |
| `projectParts` | `ProjectPart[]` | Yes | Components of the project |
| `artifacts` | `Artifact[]` | No | Registered analysis artifacts |

### ProjectPart

Defines a logical component of the project with its location and optional documentation.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Unique identifier for this project part |
| `directoryAbs` | `string` | Yes | Absolute path to the component's root directory |
| `documentationAbs` | `string` | No | Absolute path to documentation file or directory |
| `description` | `string` | No | Human-readable description |

**Constraints:**
- `name` must be unique across all project parts
- `name` must be non-empty, snake_case recommended
- `directoryAbs` must be an absolute path

### Artifact

Defines a registered analysis artifact that can be referenced by steps.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier (snake_case) |
| `description` | `string` | Yes | What this artifact captures |
| `producedBy` | `string` | Yes | `{epic}/{milestone}/{step}` that created it |
| `createdAt` | `string` | Yes | ISO date of creation (YYYY-MM-DD) |

**Constraints:**
- `id` must be unique across all artifacts
- `id` must match pattern `^[a-z][a-z0-9_]*$`
- Artifact file must exist at `.accelyst/artifacts/{id}.md`

### Registry Example

```yaml
version: 1

projectParts:
  - name: client
    directoryAbs: /home/user/project/client
    description: "Frontend React application"
  - name: api
    directoryAbs: /home/user/project/api
    documentationAbs: /home/user/project/docs/api.md
    description: "Go backend API"
  - name: test
    directoryAbs: /home/user/project/tests
    description: "Test suites"

artifacts:
  - id: auth_flow
    description: "Authentication flow including JWT handling"
    producedBy: auth-system/jwt_auth/analyze_auth
    createdAt: "2024-01-08"
```

---

## Epic Schema

Epics are ephemeral collections of related milestones. They can be regenerated without losing project knowledge.

### Root Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Human-readable epic name |
| `description` | `string` | No | What this epic accomplishes |
| `priority` | `string` | No | `ready`, `in_progress`, `blocked`, or `ideas` |
| `milestones` | `Milestone[]` | Yes | List of milestones |

### Milestone

Defines a feature or goal composed of ordered steps. Milestones can depend on other milestones within the same epic, forming a directed acyclic graph (DAG).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier within epic (snake_case) |
| `name` | `string` | Yes | Human-readable name |
| `dependsOn` | `string[]` | No | Milestone IDs that must complete first |
| `steps` | `Step[]` | Yes | Steps to complete the milestone |

**Constraints:**
- `id` must be unique within the epic
- `id` must be non-empty
- `name` must be non-empty
- `dependsOn` must reference valid milestone IDs within the same epic
- `dependsOn` must not create circular dependencies
- `steps` must contain at least one step

### Step

Defines a single action within a milestone.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier within milestone (snake_case) |
| `instruction` | `string` | Yes | Action to perform |
| `dependsOn` | `string[]` | No | Step IDs that must complete before this step |
| `projectParts` | `string[]` | No | Project part names relevant to this step |
| `produces` | `string` | No | Artifact ID this step creates |
| `requires` | `string[]` | No | Artifact IDs this step needs as context |
| `modifies` | `string[]` | No | File paths this step will modify (for concurrency safety) |

**Constraints:**
- `id` must be unique within the milestone
- `id` must be non-empty
- `dependsOn` must reference valid step IDs within the same milestone
- `dependsOn` must not create circular dependencies
- `projectParts` must reference valid project part names in registry
- `produces` must be unique across all epics and not exist in registry
- `requires` must reference artifacts in registry or `produces` from earlier steps
- `modifies` paths should be relative to project root

### Epic Example

```yaml
name: "Fullscreen Enhancements"
description: "Improve fullscreen image viewing experience"
priority: ready

milestones:
  - id: fullscreen_download
    name: "Download Button"
    dependsOn: []
    steps:
      - id: analyze_fullscreen
        instruction: "analyze current fullscreen view implementation"
        produces: fullscreen_structure
        projectParts: [client]

      - id: impl_download
        instruction: "implement download button"
        dependsOn: [analyze_fullscreen]
        projectParts: [client]
        modifies:
          - client/fullscreen-viewer.js
          - client/index.html

      - id: test_download
        instruction: "write tests for download functionality"
        dependsOn: [impl_download]
        projectParts: [test]
        modifies:
          - client/e2e/fullscreen.spec.ts

  - id: fullscreen_zoom
    name: "Zoom Controls"
    dependsOn: [fullscreen_download]
    steps:
      - id: design_zoom
        instruction: "design zoom interaction patterns"
        requires: [fullscreen_structure]
        projectParts: [client]

      - id: impl_zoom
        instruction: "implement pinch/scroll zoom"
        dependsOn: [design_zoom]
        projectParts: [client]

      - id: test_zoom
        instruction: "write tests for zoom interactions"
        dependsOn: [impl_zoom]
        projectParts: [test]
```

---

## Artifacts

Artifacts are markdown files capturing analysis outputs that persist across iterations.

### Location

```
.accelyst/artifacts/{artifact_id}.md
```

### Content

Artifacts should capture:
- Component locations and file paths
- Architecture patterns and design decisions
- Key interfaces, types, and contracts
- Extension points for future work

### Artifact Lifecycle

**Production (`produces`):**
1. Step with `produces` field executes
2. Agent writes analysis to `.accelyst/artifacts/{id}.md`
3. Registry is updated with new artifact entry

**Consumption (`requires`):**
1. Step with `requires` field is processed
2. Artifact content is injected into prompt context
3. Agent receives prior analysis as background

---

## Execution Semantics

### Milestone Dependency Resolution

Milestones form a DAG where edges represent dependencies. Accelyst performs topological sorting using Kahn's algorithm.

```
Milestone A (dependsOn: [])           → executes first
Milestone B (dependsOn: [])           → executes first (parallel candidate)
Milestone C (dependsOn: ["A", "B"])   → executes after A and B
```

### Step Dependency Resolution

Within each milestone, steps form their own DAG sorted independently.

### Artifact Dependency Resolution

Steps with `requires` must have their artifacts available:
- From registry (prior iterations)
- From `produces` in earlier step (same or dependent milestone)

### Concurrency Safety

The `modifies` field enables safe parallel execution by declaring which files a step will change.

**Conflict Detection:**
```
Step A: modifies: [app.js, config.js]
Step B: modifies: [app.js, utils.js]
Step C: modifies: [types.ts]

Conflict: A and B both modify app.js
Safe to parallel: A/B with C (no overlap)
```

**Scheduling Rules:**
1. Steps with overlapping `modifies` must execute sequentially
2. Steps with disjoint `modifies` can execute in parallel
3. Steps without `modifies` are treated as potentially modifying any file (conservative)

**When to Populate:**
- Implementation steps should always declare `modifies`
- Analysis steps (with `produces`) typically don't modify files
- Test steps should declare new test files they create

### Prompt Generation

For each step, a prompt fragment is generated:

```
<id>: [subagent prefix] [project parts clause] [requires clause] [dependencies clause] <instruction> [modifies clause]
```

**Components:**

1. **Subagent prefix** (if no step dependencies):
   ```
   use a subagent to
   ```

2. **Project parts clause** (if projectParts specified):
   ```
   Analyze <name> (<directoryAbs>) [(docs: <documentationAbs>)], ... then
   ```

3. **Requires clause** (if requires specified, injected before prompt):
   ```
   ### Context: <artifact_id>
   [artifact content]
   ```

4. **Dependencies clause** (if dependsOn specified):
   ```
   read results from steps <id>, <id>, ... then
   ```

5. **Produces suffix** (if produces specified):
   ```

   ---
   Save analysis to: .accelyst/artifacts/<id>.md
   ```

6. **Modifies clause** (if modifies specified):
   ```
   [modifies: <file1>, <file2>, ...]
   ```

7. **Instruction**: The `instruction` field value

---

## Example Output

Given the epic example above:

```
# Milestone: Download Button

## Step: analyze_fullscreen

use a subagent to Analyze client (/home/user/project/client) then analyze current fullscreen view implementation

---
Save analysis to: .accelyst/artifacts/fullscreen_structure.md
Include: component locations, state management, key interfaces, extension points

---

## Step: impl_download

Analyze client (/home/user/project/client) then read results from steps analyze_fullscreen then implement download button [modifies: client/fullscreen-viewer.js, client/index.html]

---

## Step: test_download

Analyze test (/home/user/project/tests) then read results from steps impl_download then write tests for download functionality [modifies: client/e2e/fullscreen.spec.ts]

---

# Milestone: Zoom Controls (depends on: fullscreen_download)

## Step: design_zoom

### Context: fullscreen_structure

[contents of .accelyst/artifacts/fullscreen_structure.md]

---

Analyze client (/home/user/project/client) then design zoom interaction patterns

---

## Step: impl_zoom

Analyze client (/home/user/project/client) then read results from steps design_zoom then implement pinch/scroll zoom

---

## Step: test_zoom

Analyze test (/home/user/project/tests) then read results from steps impl_zoom then write tests for zoom interactions
```

---

## CLI Interface

```bash
# Initialize project
accelyst init

# List epics
accelyst epics list

# Run milestone
accelyst run <epic>/<milestone>

# Run specific step
accelyst run <epic>/<milestone>/<step>

# Registry management
accelyst registry parts
accelyst registry artifacts
accelyst registry add-artifact <id> <description> <producedBy>
```

---

## JSON Schemas

### Registry Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "projectParts"],
  "properties": {
    "version": { "type": "integer", "const": 1 },
    "projectParts": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "directoryAbs"],
        "properties": {
          "name": { "type": "string", "minLength": 1, "pattern": "^[a-z][a-z0-9_]*$" },
          "directoryAbs": { "type": "string", "minLength": 1 },
          "documentationAbs": { "type": "string" },
          "description": { "type": "string" }
        }
      }
    },
    "artifacts": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "description", "producedBy", "createdAt"],
        "properties": {
          "id": { "type": "string", "minLength": 1, "pattern": "^[a-z][a-z0-9_]*$" },
          "description": { "type": "string", "minLength": 1 },
          "producedBy": { "type": "string", "minLength": 1 },
          "createdAt": { "type": "string", "format": "date" }
        }
      }
    }
  }
}
```

### Epic Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "milestones"],
  "properties": {
    "name": { "type": "string", "minLength": 1 },
    "description": { "type": "string" },
    "priority": { "type": "string", "enum": ["ready", "in_progress", "blocked", "ideas"] },
    "milestones": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "name", "steps"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "name": { "type": "string", "minLength": 1 },
          "dependsOn": {
            "type": "array",
            "items": { "type": "string", "minLength": 1 }
          },
          "steps": {
            "type": "array",
            "minItems": 1,
            "items": {
              "type": "object",
              "required": ["id", "instruction"],
              "properties": {
                "id": { "type": "string", "minLength": 1 },
                "instruction": { "type": "string", "minLength": 1 },
                "dependsOn": {
                  "type": "array",
                  "items": { "type": "string", "minLength": 1 }
                },
                "projectParts": {
                  "type": "array",
                  "items": { "type": "string" }
                },
                "produces": { "type": "string", "pattern": "^[a-z][a-z0-9_]*$" },
                "requires": {
                  "type": "array",
                  "items": { "type": "string", "pattern": "^[a-z][a-z0-9_]*$" }
                },
                "modifies": {
                  "type": "array",
                  "items": { "type": "string", "minLength": 1 }
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

---

## Migration from v1

Projects using single `accelyst.yaml` can migrate:

```bash
accelyst migrate
```

This will:
1. Create `.accelyst/` directory structure
2. Extract `projectParts` to `.accelyst/registry.yaml`
3. Move milestones to `.accelyst/epics/default.yaml`
4. Backup original file to `accelyst.yaml.bak`
