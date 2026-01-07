# Accelyst Configuration Specification

Version: 1.1.0

## Overview

An Accelyst configuration file defines the structure of a project and the milestones required to evolve it. The file format is YAML.

## Schema

### Root Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectParts` | `ProjectPart[]` | Yes | Components of the project |
| `milestones` | `Milestone[]` | Yes | List of milestones to execute |

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

Defines a feature or goal composed of ordered steps. Milestones can depend on other milestones, forming a directed acyclic graph (DAG) for execution ordering.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier across all milestones (alphanumeric, snake_case recommended) |
| `name` | `string` | Yes | Human-readable name for the milestone |
| `dependsOn` | `string[]` | No | Milestone IDs that must complete before this milestone |
| `steps` | `Step[]` | Yes | Steps required to complete the milestone |

**Constraints:**
- `id` must be unique across all milestones
- `id` must be a non-empty string
- `name` must be non-empty
- `dependsOn` must reference valid milestone IDs
- `dependsOn` must not create circular dependencies
- `steps` must contain at least one step

### Step

Defines a single action within a milestone.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | `string` | Yes | Unique identifier within the milestone (alphanumeric, snake_case recommended) |
| `instruction` | `string` | Yes | Action to perform |
| `dependsOn` | `string[]` | No | Step IDs that must complete before this step |
| `projectParts` | `string[]` | No | Project part names relevant to this step |
| `acceptanceCriteria` | `string[]` | No | Conditions that define successful completion |
| `tier` | `string` | No | Execution tier: `ai` (autonomous) or `human` (requires human). Defaults to `ai` |

**Constraints:**
- `id` must be unique within the milestone
- `id` must be a non-empty string
- `dependsOn` must reference valid step IDs within the same milestone
- `dependsOn` must not create circular dependencies
- `projectParts` must reference valid project part names
- `tier` must be one of: `ai`, `human`

**Tier Semantics:**
- `ai`: Step can be fully executed by an AI agent autonomously
- `human`: Step requires human action (e.g., approvals, external account setup, physical tasks)

## Execution Semantics

### Milestone Dependency Resolution

Milestones form a directed acyclic graph (DAG) where edges represent dependencies. Accelyst performs topological sorting using Kahn's algorithm to determine milestone execution order.

```
Milestone A (id: "auth", dependsOn: [])              → executes first (no dependencies)
Milestone B (id: "db", dependsOn: [])                → executes first (no dependencies)
Milestone C (id: "api", dependsOn: ["auth", "db"])   → executes after A and B complete
```

Milestones with no dependencies (`dependsOn` is empty or omitted) are candidates for parallel execution.

### Step Dependency Resolution

Within each milestone, steps form their own DAG. Steps are topologically sorted independently per milestone.

```
Step A (id: "research", dependsOn: [])               → executes first (no dependencies)
Step B (id: "plan", dependsOn: [])                   → executes first (no dependencies)
Step C (id: "impl", dependsOn: ["research", "plan"]) → executes after A and B complete
```

### Parallel Execution

Both milestones and steps with no dependencies are candidates for parallel execution. These receive the `use a subagent to` prefix in generated prompts.

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

3. **Dependencies clause** (if dependsOn specified):
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
  - id: jwt_infrastructure
    name: "JWT Infrastructure"
    dependsOn: []
    steps:
      - id: research_jwt
        instruction: "research JWT best practices"
        dependsOn: []
        projectParts: []
        acceptanceCriteria:
          - "documented token expiry strategy"
          - "documented refresh token approach"
      - id: impl_jwt_service
        instruction: "implement JWT service"
        dependsOn:
          - research_jwt
        projectParts:
          - api
        acceptanceCriteria:
          - "JWT generation and validation working"
          - "unit tests passing"
        tier: ai

  - id: session_management
    name: "Session Management"
    dependsOn: []
    steps:
      - id: analyze_sessions
        instruction: "analyze existing session patterns"
        dependsOn: []
        projectParts:
          - api
      - id: impl_sessions
        instruction: "implement session handling"
        dependsOn:
          - analyze_sessions
        projectParts:
          - api

  - id: login_ui
    name: "Login UI"
    dependsOn:
      - jwt_infrastructure
      - session_management
    steps:
      - id: impl_login_form
        instruction: "implement login form component"
        dependsOn: []
        projectParts:
          - client
        acceptanceCriteria:
          - "form renders with email and password fields"
          - "validation errors display correctly"
      - id: impl_auth_flow
        instruction: "implement authentication flow"
        dependsOn:
          - impl_login_form
        projectParts:
          - client
        acceptanceCriteria:
          - "login redirects to dashboard on success"
          - "error states handled gracefully"
      - id: configure_oauth_provider
        instruction: "configure OAuth provider credentials in production"
        dependsOn:
          - impl_auth_flow
        projectParts: []
        tier: human
        acceptanceCriteria:
          - "OAuth client ID and secret configured"
          - "redirect URIs registered"
```

**Output:**

```
# Milestone: JWT Infrastructure

research_jwt: use a subagent to research JWT best practices
impl_jwt_service: Analyze api (/home/user/project/api) (docs: /home/user/project/docs/api.md) then read results from steps research_jwt then implement JWT service

---

# Milestone: Session Management

analyze_sessions: use a subagent to Analyze api (/home/user/project/api) (docs: /home/user/project/docs/api.md) then analyze existing session patterns
impl_sessions: Analyze api (/home/user/project/api) (docs: /home/user/project/docs/api.md) then read results from steps analyze_sessions then implement session handling

---

# Milestone: Login UI (depends on: jwt_infrastructure, session_management)

impl_login_form: use a subagent to Analyze client (/home/user/project/client) then implement login form component
impl_auth_flow: Analyze client (/home/user/project/client) then read results from steps impl_login_form then implement authentication flow
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
                "acceptanceCriteria": {
                  "type": "array",
                  "items": { "type": "string", "minLength": 1 }
                },
                "tier": {
                  "type": "string",
                  "enum": ["ai", "human"],
                  "default": "ai"
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
