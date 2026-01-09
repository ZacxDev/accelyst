# Registry and Epics Architecture

> Feature specification for persistent project knowledge and multi-epic support

## Overview

This document outlines architectural changes to separate persistent project knowledge (registry, artifacts) from ephemeral work items (epics, milestones). This enables:

- **Knowledge accumulation** across project iterations
- **Multiple concurrent epics** for parallel workstreams
- **Disposable milestone configs** that can be regenerated
- **Cross-epic artifact sharing** for building on prior analysis

---

## File Structure

```
project/
├── .accelyst/
│   ├── registry.yaml           # Persistent project knowledge
│   ├── epics/                   # Ephemeral work items
│   │   ├── fullscreen-enhancements.yaml
│   │   ├── codemirror-ide.yaml
│   │   └── admin-system.yaml
│   └── artifacts/               # Produced analysis outputs
│       ├── fullscreen_structure.md
│       ├── auth_flow.md
│       └── codemirror_extensions.md
```

---

## Registry (`registry.yaml`)

The registry is the persistent source of truth for project-level definitions.

### Schema

```yaml
# .accelyst/registry.yaml

version: 1

projectParts:
  - name: client
    directoryAbs: /home/user/project/client
    description: "Frontend React application"

  - name: api
    directoryAbs: /home/user/project/api
    description: "Go backend API"

  - name: dspy-service
    directoryAbs: /home/user/project/dspy-service
    description: "Python DSPy inference service"

  - name: test
    directoryAbs: /home/user/project/tests
    description: "Test suites"

artifacts:
  - id: fullscreen_structure
    description: "Fullscreen view component architecture and state management"
    producedBy: fullscreen-enhancements/fullscreen_download/analyze_fullscreen
    createdAt: 2024-01-08

  - id: auth_flow
    description: "Authentication flow including JWT handling and session management"
    producedBy: admin-system/admin_auth/analyze_auth
    createdAt: 2024-01-05

  - id: codemirror_extensions
    description: "CodeMirror extension architecture and custom language support"
    producedBy: codemirror-ide/syntax_highlighting/analyze_codemirror
    createdAt: 2024-01-10
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | integer | Yes | Schema version for migrations |
| `projectParts` | array | Yes | Logical components of the project |
| `projectParts[].name` | string | Yes | Identifier (snake_case) |
| `projectParts[].directoryAbs` | string | Yes | Absolute path to component |
| `projectParts[].description` | string | No | Human-readable description |
| `artifacts` | array | No | Registered analysis artifacts |
| `artifacts[].id` | string | Yes | Unique identifier (snake_case) |
| `artifacts[].description` | string | Yes | What this artifact captures |
| `artifacts[].producedBy` | string | Yes | `{epic}/{milestone}/{step}` that created it |
| `artifacts[].createdAt` | date | Yes | ISO date of creation |

---

## Epics (`epics/*.yaml`)

Epics are ephemeral collections of related milestones. They can be:
- Generated from planning briefs
- Created manually
- Regenerated/discarded without losing project knowledge

### Schema

```yaml
# .accelyst/epics/fullscreen-enhancements.yaml

name: "Fullscreen View Enhancements"
description: "Improve fullscreen image viewing with download, zoom, and navigation"
priority: ready  # ready | in_progress | blocked | ideas

milestones:
  - id: fullscreen_download
    name: "Fullscreen Download Button"
    dependsOn: []
    steps:
      - id: analyze_fullscreen
        instruction: "analyze current fullscreen view component implementation"
        produces: fullscreen_structure
        projectParts: [client]

      - id: impl_download
        instruction: "implement download button that saves current image"
        dependsOn: [analyze_fullscreen]
        projectParts: [client]

      - id: test_download
        instruction: "write tests for download functionality"
        dependsOn: [impl_download]
        projectParts: [test]

  - id: fullscreen_zoom
    name: "Fullscreen Zoom Controls"
    dependsOn: [fullscreen_download]
    steps:
      - id: design_zoom
        instruction: "design zoom page UX and interaction patterns"
        requires: [fullscreen_structure]  # From prior milestone
        projectParts: [client]

      - id: impl_zoom
        instruction: "implement pinch/scroll zoom with tap/drag pan"
        dependsOn: [design_zoom]
        projectParts: [client]

      - id: test_zoom
        instruction: "write tests for zoom interactions"
        dependsOn: [impl_zoom]
        projectParts: [test]
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable epic name |
| `description` | string | No | What this epic accomplishes |
| `priority` | enum | No | `ready`, `in_progress`, `blocked`, `ideas` |
| `milestones` | array | Yes | Ordered list of milestones |

#### Milestone Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique within epic (snake_case) |
| `name` | string | Yes | Human-readable name |
| `dependsOn` | array | No | Milestone IDs that must complete first |
| `steps` | array | Yes | Ordered list of steps |

#### Step Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique within milestone (snake_case) |
| `instruction` | string | Yes | What to do |
| `dependsOn` | array | No | Step IDs within this milestone |
| `projectParts` | array | No | Components to analyze |
| `produces` | string | No | Artifact ID this step creates |
| `requires` | array | No | Artifact IDs this step needs |

---

## Artifacts

Artifacts are markdown files capturing analysis outputs that can be reused across milestones and epics.

### Location

```
.accelyst/artifacts/{artifact_id}.md
```

### Content Guidelines

Artifacts should capture:
- Component locations and file paths
- Architecture patterns and design decisions
- Key interfaces, types, and contracts
- Extension points for future work
- Dependencies and relationships

### Example Artifact

```markdown
# fullscreen_structure

> Produced by: fullscreen-enhancements/fullscreen_download/analyze_fullscreen
> Created: 2024-01-08

## Component Location

- Main component: `src/components/Fullscreen/index.tsx`
- Styles: `src/components/Fullscreen/styles.module.css`
- Hook: `src/hooks/useFullscreen.ts`

## State Management

Uses Zustand store at `src/stores/fullscreenStore.ts`:
- `isOpen: boolean`
- `currentImage: Image | null`
- `zoomLevel: number`

## Key Interfaces

```typescript
interface FullscreenProps {
  imageId: string;
  onClose: () => void;
}
```

## Extension Points

- Add toolbar buttons in `FullscreenToolbar` component
- Zoom controls would integrate with `zoomLevel` state
- Download can use `currentImage.url` directly
```

---

## Artifact Lifecycle

### Production (`produces`)

When a step with `produces` completes:

1. **Agent writes artifact** to `.accelyst/artifacts/{id}.md`
2. **Accelyst updates registry** with new artifact entry:
   ```yaml
   artifacts:
     - id: fullscreen_structure
       description: (from artifact header or instruction)
       producedBy: fullscreen-enhancements/fullscreen_download/analyze_fullscreen
       createdAt: 2024-01-08
   ```

### Consumption (`requires`)

When generating a prompt for a step with `requires`:

1. **Validate** artifact exists in registry
2. **Read** `.accelyst/artifacts/{id}.md`
3. **Inject** content into prompt context

### Prompt Generation

**Step with `produces`:**
```
## Step: analyze_fullscreen

use a subagent to Analyze client (/path/to/client) then analyze current fullscreen view component implementation

---
📝 Save analysis to: .accelyst/artifacts/fullscreen_structure.md
Include: component locations, state management, key interfaces, extension points
```

**Step with `requires`:**
```
## Step: impl_zoom

### Required Context

#### fullscreen_structure (from fullscreen_download/analyze_fullscreen)

[contents of .accelyst/artifacts/fullscreen_structure.md]

---

implement pinch/scroll zoom with tap/drag pan controls
```

---

## Validation Rules

### Registry Validation

| Rule | Error |
|------|-------|
| `projectParts[].name` must be unique | "Duplicate project part: {name}" |
| `projectParts[].directoryAbs` must exist | "Directory not found: {path}" |
| `artifacts[].id` must be unique | "Duplicate artifact: {id}" |
| Artifact file must exist | "Artifact file missing: {id}.md" |

### Epic Validation

| Rule | Error |
|------|-------|
| `milestones[].id` unique within epic | "Duplicate milestone: {id}" |
| `milestones[].dependsOn` references valid IDs | "Unknown milestone dependency: {id}" |
| `steps[].id` unique within milestone | "Duplicate step: {id}" |
| `steps[].dependsOn` references valid step IDs | "Unknown step dependency: {id}" |
| `steps[].projectParts` in registry | "Unknown project part: {name}" |
| `steps[].requires` in registry or produced earlier | "Unknown artifact: {id}" |
| `steps[].produces` not already in registry | "Artifact already exists: {id}" |
| No circular dependencies | "Circular dependency detected: {path}" |

### Cross-Epic Validation

| Rule | Error |
|------|-------|
| `produces` unique across all epics | "Artifact {id} produced by multiple steps" |

---

## CLI Interface

### Initialization

```bash
# Initialize registry in current directory
accelyst init

# Creates:
#   .accelyst/registry.yaml (with prompts for project parts)
#   .accelyst/epics/
#   .accelyst/artifacts/
```

### Running Epics

```bash
# List all epics and their status
accelyst epics list

# Run specific milestone from an epic
accelyst run fullscreen-enhancements/fullscreen_download

# Run all ready milestones in an epic
accelyst run fullscreen-enhancements --all

# Run with specific step
accelyst run fullscreen-enhancements/fullscreen_download/analyze_fullscreen
```

### Registry Management

```bash
# List project parts
accelyst registry parts

# List artifacts
accelyst registry artifacts

# Add artifact manually (from prior analysis)
accelyst registry add-artifact auth_flow \
  --description "Auth flow analysis" \
  --file ./my-analysis.md

# Validate registry
accelyst registry validate
```

### Epic Management

```bash
# Create new epic from template
accelyst epic new my-feature

# Validate epic
accelyst epic validate fullscreen-enhancements

# Archive completed epic
accelyst epic archive fullscreen-enhancements
```

---

## Migration from Single Config

Projects using single `accelyst.yaml` can migrate:

```bash
accelyst migrate
```

This will:
1. Extract `projectParts` to `.accelyst/registry.yaml`
2. Move milestones to `.accelyst/epics/default.yaml`
3. Rename original file to `accelyst.yaml.bak`

---

## Example Workflow

### 1. Initialize Project

```bash
cd /home/user/promptver
accelyst init
# Answer prompts for project parts
```

### 2. Create Epic from Planning Brief

```bash
# Use milestone-plan skill to create planning brief
# Then convert to epic
accelyst epic create-from-brief planning-brief.yaml --name fullscreen-enhancements
```

### 3. Run First Milestone

```bash
accelyst run fullscreen-enhancements/fullscreen_download
# Agent executes steps
# analyze_fullscreen produces fullscreen_structure artifact
# Registry updated automatically
```

### 4. Continue with Dependent Work

```bash
accelyst run fullscreen-enhancements/fullscreen_zoom
# impl_zoom step receives fullscreen_structure artifact in context
```

### 5. Start New Epic Using Existing Artifacts

```bash
accelyst epic new image-gallery

# Edit .accelyst/epics/image-gallery.yaml
# Add step that requires: [fullscreen_structure]
# Works because artifact already in registry

accelyst run image-gallery/gallery_view
```

---

## Benefits

| Before | After |
|--------|-------|
| Single config file | Multiple concurrent epics |
| Project parts repeated | Project parts defined once |
| Analysis redone each milestone | Artifacts persist and accumulate |
| Config is precious | Epics are disposable |
| Linear milestone execution | Parallel epic progress |

---

## Future Considerations

### Artifact Versioning
Track multiple versions of an artifact as understanding evolves:
```yaml
artifacts:
  - id: auth_flow
    version: 2
    supersedes: auth_flow_v1
```

### Artifact Staleness
Detect when artifacts may be outdated:
- Track git commit SHA when created
- Warn if referenced files changed since creation

### Epic Templates
Reusable epic patterns for common workflows:
```bash
accelyst epic new my-api --template crud-api
```

### Artifact Search
Find artifacts by content or metadata:
```bash
accelyst artifacts search "zustand"
```
