# Artifact Management

This document details how the milestone-execute skill handles artifact production and consumption.

## Artifact Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                     PRODUCTION                               │
│  Step with produces: auth_analysis                          │
│                                                             │
│  1. Subagent executes analysis                              │
│  2. Subagent writes to .accelyst/artifacts/auth_analysis.md │
│  3. Orchestrator verifies file exists                       │
│  4. Orchestrator updates registry                           │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     STORAGE                                  │
│  .accelyst/artifacts/auth_analysis.md                       │
│  .accelyst/registry.yaml (artifacts section)                │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    CONSUMPTION                               │
│  Step with requires: [auth_analysis]                        │
│                                                             │
│  1. Orchestrator reads artifact content                     │
│  2. Content injected into subagent prompt                   │
│  3. Subagent receives full context                          │
└─────────────────────────────────────────────────────────────┘
```

## File Locations

| Item | Path |
|------|------|
| Artifact files | `.accelyst/artifacts/{artifact_id}.md` |
| Registry | `.accelyst/registry.yaml` |

## Production Process

### 1. Step Definition

```yaml
steps:
  - id: analyze_auth
    instruction: "analyze existing authentication patterns"
    produces: auth_analysis
    projectParts:
      - api
```

### 2. Subagent Prompt

The subagent receives instructions to create the artifact:

```markdown
## Output Required

Save your analysis to: `.accelyst/artifacts/auth_analysis.md`

Structure your analysis as follows:
[artifact format guidance from EXECUTION.md]
```

### 3. Subagent Creates Artifact

The subagent writes to the specified path using the Write tool.

### 4. Verification

After subagent completes, orchestrator verifies:

```
if not file_exists(".accelyst/artifacts/auth_analysis.md"):
    # Handle missing artifact
    AskUserQuestion: "Artifact auth_analysis was not created. Options:
    1. Retry the step
    2. Create it manually
    3. Skip and continue"
```

### 5. Registry Update

Add artifact to registry:

```yaml
# .accelyst/registry.yaml
artifacts:
  - id: auth_analysis
    description: "Analysis of authentication patterns and flow"
    producedBy: feature-auth/auth_api/analyze_auth
    createdAt: "2025-01-09"
```

## Consumption Process

### 1. Step Definition

```yaml
steps:
  - id: impl_auth
    instruction: "implement authentication endpoints"
    requires:
      - auth_analysis
    dependsOn:
      - analyze_auth
    projectParts:
      - api
```

### 2. Pre-execution Check

Before spawning subagent:

```
for artifact_id in step.requires:
    path = f".accelyst/artifacts/{artifact_id}.md"
    if not file_exists(path):
        # Check if another step produces it
        producer = find_producer(artifact_id)
        if producer and not producer.completed:
            error("Dependency not met: {artifact_id} not yet produced")
        else:
            error("Artifact {artifact_id} not found")
```

### 3. Content Injection

Read artifact and inject into prompt:

```markdown
## Required Context

### Artifact: auth_analysis

# auth_analysis

## Overview
The authentication system uses JWT tokens with refresh token rotation...

## Key Files
| File | Purpose |
|------|---------|
| api/auth.go | Main authentication handlers |
| api/middleware.go | JWT validation middleware |

[... full artifact content ...]

---
```

### 4. Subagent Execution

Subagent receives full artifact content as part of prompt, no need to read files.

## Artifact Format

### Recommended Structure

```markdown
# {artifact_id}

## Overview
Brief 2-3 sentence summary of the analysis.

## Key Files
| File | Purpose |
|------|---------|
| path/to/file.ext | What this file does |

## Patterns Identified
- **Pattern Name**: Description of the pattern
- **Pattern Name**: Description of the pattern

## Key Interfaces

```typescript
// Important types or interfaces
interface AuthConfig {
  jwtSecret: string;
  tokenExpiry: number;
}
```

## Architecture

```
Component A ──► Component B ──► Component C
     │              │
     ▼              ▼
  Database      External API
```

## State Management
Description of how state flows through the system.

## Extension Points
- Where to add new authentication providers
- How to extend the token validation

## Recommendations
1. Specific suggestion for implementation
2. Another suggestion
```

### Content Guidelines

| Do | Don't |
|----|-------|
| Include specific file paths | Use vague references |
| Show code snippets for patterns | Include entire files |
| Note line numbers for key code | Assume reader knows the codebase |
| Explain why patterns exist | Only describe what exists |
| Recommend specific approaches | Leave decisions ambiguous |

## Error Handling

### Missing Required Artifact

```
Step impl_auth requires artifact 'auth_analysis' which doesn't exist.

Possible causes:
1. analyze_auth step hasn't run yet
2. analyze_auth step failed to produce the artifact
3. Artifact file was deleted

Options:
1. Execute analyze_auth step first
2. Create the artifact manually
3. Skip this requirement (implementation may fail)
```

### Artifact Not Produced

```
Step analyze_auth completed but artifact 'auth_analysis' was not created.

The subagent should have written to:
  .accelyst/artifacts/auth_analysis.md

Options:
1. Retry the step
2. Check subagent output for errors
3. Create artifact manually
4. Mark step as failed and continue
```

### Corrupt or Invalid Artifact

```
Artifact 'auth_analysis' exists but appears to be empty or invalid.

File: .accelyst/artifacts/auth_analysis.md
Size: 0 bytes

Options:
1. Retry the producing step
2. Edit artifact manually
3. Delete and regenerate
```

## Cross-Milestone Artifacts

Artifacts can be shared across milestones:

```yaml
# Milestone 1
- id: analyze_api
  produces: api_architecture

# Milestone 2 (depends on Milestone 1)
- id: impl_feature
  requires:
    - api_architecture  # Uses artifact from Milestone 1
```

The orchestrator:
1. Checks registry for existing artifacts
2. Checks current execution plan for pending producers
3. Validates dependency ordering

## Registry Schema

```yaml
# .accelyst/registry.yaml

version: 1

projectParts:
  - name: api
    directoryAbs: /path/to/api
    description: "Backend API"

artifacts:
  - id: auth_analysis
    description: "Analysis of authentication flow"
    producedBy: feature-auth/auth_api/analyze_auth
    createdAt: "2025-01-09"

  - id: api_architecture
    description: "Overall API architecture patterns"
    producedBy: onboarding/initial_analysis/analyze_api
    createdAt: "2025-01-05"
```

### Registry Update Logic

After artifact production:

```python
def update_registry_artifact(artifact_id, step, registry_path):
    # Read current registry
    registry = yaml.load(read_file(registry_path))

    # Check if artifact already exists
    existing = find_artifact(registry, artifact_id)
    if existing:
        # Update existing entry
        existing.producedBy = f"{epic}/{milestone}/{step.id}"
        existing.createdAt = today()
    else:
        # Add new entry
        registry.artifacts.append({
            "id": artifact_id,
            "description": infer_description(artifact_id),
            "producedBy": f"{epic}/{milestone}/{step.id}",
            "createdAt": today()
        })

    # Write back
    write_file(registry_path, yaml.dump(registry))
```

## Artifact Discovery

### Finding Available Artifacts

```python
def get_available_artifacts():
    # 1. Check registry
    registry = load_registry()
    registered = registry.artifacts

    # 2. Scan artifacts directory
    files = glob(".accelyst/artifacts/*.md")
    on_disk = [file.stem for file in files]

    # 3. Check current execution plan
    pending = [step.produces for step in plan if step.produces]

    return {
        "registered": registered,
        "on_disk": on_disk,
        "pending": pending,
        "available": set(registered) | set(on_disk),
        "orphaned": set(on_disk) - set(registered)
    }
```

### Orphaned Artifact Warning

If artifacts exist on disk but not in registry:

```
Warning: Found orphaned artifacts (not in registry):
  - .accelyst/artifacts/old_analysis.md
  - .accelyst/artifacts/temp_notes.md

These may be from previous iterations. Options:
1. Register them (provide description and producedBy)
2. Delete them
3. Ignore for now
```

## Cleanup and Maintenance

### Stale Artifact Detection

Artifacts may become stale if:
- The producing step's projectParts changed significantly
- The codebase evolved since production
- The artifact references deleted files

Suggestion: Include timestamp checking:

```python
def check_artifact_freshness(artifact_id, max_age_days=30):
    registry_entry = find_artifact(registry, artifact_id)
    if not registry_entry:
        return "unregistered"

    age = today() - parse_date(registry_entry.createdAt)
    if age.days > max_age_days:
        return "stale"

    return "fresh"
```

### Manual Artifact Creation

Users can create artifacts manually:

```bash
# Create artifact file
cat > .accelyst/artifacts/my_analysis.md << 'EOF'
# my_analysis

## Overview
Manual analysis of the component...
EOF

# Register in registry
# (Edit .accelyst/registry.yaml to add entry)
```

Or use the accelyst CLI:

```bash
./accelyst registry add-artifact my_analysis "Description of analysis" manual/creation/step
```
