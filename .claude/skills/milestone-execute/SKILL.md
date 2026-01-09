---
name: milestone-execute
description: Execute Accelyst milestones by orchestrating subagents for each step. Use when running "accelyst run", executing milestones, or when the user wants to implement features from an epic. Handles dependency resolution, artifact injection, and parallel execution.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, Task, TodoWrite, AskUserQuestion
---

# Milestone Execution Orchestrator

Execute Accelyst milestones by spawning focused subagents for each step.

> **Dependency**: This skill uses `/accelyst-guide` for CLI operations and architecture reference.
> See [accelyst-guide](../accelyst-guide/SKILL.md) for command patterns and configuration management.

## Architecture

```
User: "/milestone-execute feature-x/auth_api"
          │
          ▼
┌─────────────────────────────┐
│  milestone-execute skill    │
│  ─────────────────────────  │
│  1. Load registry + epic    │
│  2. Resolve dependencies    │
│  3. Build execution DAG     │
│  4. Create TodoWrite items  │
└─────────────────────────────┘
          │
          ▼
┌─────────────────────────────┐
│   For each ready step:      │
│   ─────────────────────────  │
│   • Load required artifacts │
│   • Build focused prompt    │
│   • Spawn subagent (Task)   │
│   • Capture output          │
│   • Save produced artifacts │
│   • Mark todo complete      │
└─────────────────────────────┘
          │
          ▼
     [Repeat until all steps complete]
```

## Invocation

```bash
# Execute entire milestone
/milestone-execute <epic>/<milestone>

# Execute single step
/milestone-execute <epic>/<milestone>/<step>

# Execute all ready milestones in an epic
/milestone-execute <epic>
```

## Process

### Phase 0: Pre-flight Checks

Use accelyst CLI to validate the project state before execution:

```bash
# Ensure binary is available (build if needed)
go build -o accelyst

# Verify project is initialized
ls .accelyst/registry.yaml || ./accelyst init

# List available epics to confirm target exists
./accelyst epics list

# Validate registry has required parts
./accelyst registry parts
```

See [accelyst-guide](../accelyst-guide/SKILL.md) for detailed CLI patterns.

### Phase 1: Load and Validate

1. **Read registry** (`.accelyst/registry.yaml`)
   ```
   Read .accelyst/registry.yaml
   → Extract projectParts with directoryAbs paths
   → Load existing artifacts list
   ```

2. **Read epic** (`.accelyst/epics/<epic>.yaml`)
   ```
   Read .accelyst/epics/{epic}.yaml
   → Parse milestone structure
   → Validate projectParts references exist in registry
   → Check artifact dependencies can be resolved
   ```

3. **Get CLI prompt format** (optional, for reference)
   ```bash
   # See what accelyst would generate for this milestone
   ./accelyst run <epic>/<milestone>
   ```

   This shows the prompt format including dependency chains and artifact injection.

4. **Resolve target**
   - If `<epic>/<milestone>/<step>`: Execute single step
   - If `<epic>/<milestone>`: Execute all steps in milestone
   - If `<epic>`: Execute all milestones in topological order

### Phase 2: Build Execution Plan

1. **Topological sort milestones** by `dependsOn` field
2. **For each milestone, topological sort steps** by `dependsOn`
3. **Identify parallel groups** - steps with no mutual dependencies
4. **Create TodoWrite items** for visibility

```
TodoWrite items:
- [ ] auth_api/analyze_auth (produces: auth_analysis)
- [ ] auth_api/implement (requires: auth_analysis)
- [ ] auth_api/test_auth
```

### Phase 3: Execute Steps

For each step in execution order:

#### 3a. Check Readiness

Before executing a step, verify:
- All steps in `dependsOn` are marked complete
- All artifacts in `requires` exist at `.accelyst/artifacts/{id}.md`

If not ready, skip and process other ready steps first.

#### 3b. Build Step Context

Assemble a focused prompt for the subagent. See [EXECUTION.md](EXECUTION.md) for templates.

**Context includes:**
- Project part paths from registry
- Full content of required artifacts
- The step instruction
- Output requirements if `produces` is specified

#### 3c. Select Subagent Type

| Step Pattern | Subagent Type | Rationale |
|--------------|---------------|-----------|
| Has `produces` field | `Explore` | Analysis/research focus |
| `test` in projectParts | `quality-engineer` | Testing expertise |
| `api` in projectParts | `backend-architect` | Backend focus |
| `client` in projectParts | `frontend-architect` | Frontend focus |
| `infra` in projectParts | `devops-architect` | Infrastructure focus |
| `security` in instruction | `security-engineer` | Security focus |
| Default | `general-purpose` | Full capabilities |

#### 3d. Spawn Subagent

```
Task(
  subagent_type: <selected_type>,
  prompt: <assembled_context>,
  description: "Execute {step.id}"
)
```

For parallel execution of independent steps:
```
Task(step_A, run_in_background: true)  → task_id_a
Task(step_B, run_in_background: true)  → task_id_b

TaskOutput(task_id_a, block: true)
TaskOutput(task_id_b, block: true)
```

#### 3e. Capture Results

After subagent completes:
1. If `produces`: Verify artifact file exists
2. Update TodoWrite status to complete
3. If failed: Offer retry, skip, or abort options

### Phase 4: Artifact Management

See [ARTIFACTS.md](ARTIFACTS.md) for detailed handling.

**Production** (`produces` field):
- Subagent writes to `.accelyst/artifacts/{id}.md`
- Verify file exists after completion
- Register artifact using CLI:
  ```bash
  ./accelyst registry add-artifact {id} "{description}" {epic}/{milestone}/{step}
  ```

**Consumption** (`requires` field):
- Check artifact exists: `./accelyst registry artifacts`
- Read artifact content before spawning subagent
- Inject full content into prompt under `## Required Context`

### Phase 5: Completion

After all steps complete:
1. Verify all TodoWrite items are marked complete
2. Update registry with any new artifacts
3. Report execution summary

## Error Handling

See [accelyst-guide Error Handling](../accelyst-guide/SKILL.md#error-handling) for CLI-based recovery patterns.

### Step Failure

```
Step {step.id} failed:
Error: {error_message}

Options:
1. Retry this step
2. Skip and continue (dependent steps may fail)
3. Abort milestone execution
4. Debug interactively
```

Use AskUserQuestion to get user's choice.

### Missing Artifact

```
Step {step.id} requires artifact '{artifact_id}' which doesn't exist.

Expected location: .accelyst/artifacts/{artifact_id}.md
Should be produced by: {producedBy from registry}

Options:
1. Execute the producing step first
2. Create artifact manually
3. Skip this requirement (may cause issues)
```

**Recovery via CLI:**
```bash
# Check what artifacts exist
./accelyst registry artifacts

# Add artifact manually if file exists
./accelyst registry add-artifact {artifact_id} "description" manual/creation
```

### Missing Project Part

If a step references a projectPart not in registry:
```
Step {step.id} references unknown project part: '{part_name}'

Add it to .accelyst/registry.yaml or fix the step.
```

**Recovery via CLI:**
```bash
# Check current parts
./accelyst registry parts

# Edit registry to add missing part
# (use Read/Write tools on .accelyst/registry.yaml)
```

Abort execution until registry is fixed.

### No .accelyst Directory

If project not initialized:
```bash
./accelyst init
```

Then populate registry with project parts before retrying.

## Execution Modes

### Interactive (Default)

- Show plan before executing
- Pause after each step for review
- Allow skipping or modifying steps

### Autonomous (`--auto`)

- Execute all steps without pausing
- Only stop on errors
- Suitable for well-tested milestones

### Dry Run (`--dry-run`)

- Show execution plan
- Don't spawn any subagents
- Validate all dependencies

## Progress Display

Use TodoWrite for real-time progress:

```
Executing milestone: auth_api

[completed] analyze_auth - Produced: auth_analysis
[in_progress] implement - Running...
[pending] test_auth - Waiting for: implement
```

## Example Session

**User:** `/milestone-execute ready-features/fullscreen_download`

**Phase 1: Load**
```
Reading registry: .accelyst/registry.yaml
→ 8 project parts found
→ 0 existing artifacts

Reading epic: .accelyst/epics/ready-features.yaml
→ 13 milestones found
→ Target: fullscreen_download
```

**Phase 2: Plan**
```
Execution Plan for: fullscreen_download

Step 1: analyze_fullscreen
  projectParts: [client]
  produces: fullscreen_view_analysis
  dependsOn: []
  → Can run immediately

Step 2: impl_download_btn
  projectParts: [client]
  requires: [fullscreen_view_analysis]
  dependsOn: [analyze_fullscreen]
  → Waits for Step 1

Step 3: test_download
  projectParts: [test]
  dependsOn: [impl_download_btn]
  → Waits for Step 2

[TodoWrite: 3 items created]
```

**Phase 3: Execute**
```
Executing Step 1: analyze_fullscreen
  Subagent: Explore (has produces field)
  Context: client at /home/user/project/client
  [Task spawned...]

  ✓ Complete - artifact saved

Executing Step 2: impl_download_btn
  Subagent: frontend-architect (client projectPart)
  Context: client + fullscreen_view_analysis artifact
  [Task spawned...]

  ✓ Complete - download button implemented

Executing Step 3: test_download
  Subagent: quality-engineer (test projectPart)
  Context: test at /home/user/project/client/e2e
  [Task spawned...]

  ✓ Complete - tests passing
```

**Phase 5: Summary**
```
Milestone complete: fullscreen_download

Steps: 3/3 executed
Artifacts produced: 1 (fullscreen_view_analysis)
Registry updated: Yes

Next milestone in epic: model_selection_ux
Run: /milestone-execute ready-features/model_selection_ux
```

## Parallel Execution Example

When steps have no dependencies on each other:

```
Execution Plan for: generator_and_ui_polish

Parallel Group 1 (no dependencies):
  - generator_optimize_history/analyze_optimize_call
  - generator_ui_polish/analyze_generator_ui
  - instruction_btn_polish/analyze_instruction_btn

Sequential (after group 1):
  - generator_optimize_history/impl_history_limit
  - generator_ui_polish/hide_start_when_generating
  - instruction_btn_polish/reduce_padding
  ...

Executing Parallel Group 1:
  [Task: analyze_optimize_call, background: true] → task_1
  [Task: analyze_generator_ui, background: true] → task_2
  [Task: analyze_instruction_btn, background: true] → task_3

  Waiting for all...
  ✓ task_1 complete (produced: optimize_call_analysis)
  ✓ task_2 complete (produced: generator_ui_analysis)
  ✓ task_3 complete (produced: instruction_btn_analysis)

All analysis steps complete. Proceeding to implementation...
```

## Quality Checks

**Before execution:**
- [ ] Registry file exists and parses correctly
- [ ] Epic file exists and parses correctly
- [ ] All projectParts in steps exist in registry
- [ ] No circular dependencies in step DAG
- [ ] All `requires` artifacts exist or will be produced

**After each step:**
- [ ] Subagent completed without critical error
- [ ] Produced artifacts exist (if applicable)
- [ ] TodoWrite item marked complete

**After milestone:**
- [ ] All steps marked complete
- [ ] All expected artifacts exist
- [ ] Registry updated with new artifacts

## Integration

### With Accelyst CLI

This skill leverages the accelyst CLI for validation and prompt generation.
See [accelyst-guide](../accelyst-guide/SKILL.md) for complete CLI reference.

**Key CLI Commands Used:**

| Command | Purpose |
|---------|---------|
| `./accelyst epics list` | Validate epic exists |
| `./accelyst registry parts` | Check project parts |
| `./accelyst registry artifacts` | Check available artifacts |
| `./accelyst run <path>` | Get reference prompt format |
| `./accelyst registry add-artifact` | Register produced artifacts |

**Using CLI Output as Reference:**
```bash
# See the prompt accelyst would generate
./accelyst run ready-features/fullscreen_download
```

This shows:
- Project part paths to analyze
- Artifact content injection
- Dependency chains
- Save instructions for `produces`

The skill replicates this format when building subagent prompts.

### With Other Skills

| Skill | Purpose | When to Use |
|-------|---------|-------------|
| `/milestone-orchestrate` | **Full-cycle automation** | Backlog → implemented features (runs all skills) |
| `/accelyst-guide` | CLI operations, config management | Troubleshooting, manual operations |
| `/milestone-plan` | Interactive backlog discovery | Vague requirements → planning brief |
| `/milestone-convert` | Convert backlogs to epic YAML | Planning brief → epic file |
| `/milestone-execute` | Orchestrate execution (this skill) | Epic file → implemented features |

**Complete Workflow:**
```
                    /milestone-orchestrate
                    (automates full chain)
                            │
Raw backlog ────────────────┼─────────────────────────────►
        │                   │                              │
        ▼                   ▼                              ▼
  /milestone-plan → Planning brief                 Implemented
        │                   │                        features
        ▼                   ▼
  /milestone-convert → Epic YAML
        │                   │
        ▼                   ▼
  /milestone-execute → Subagent execution
```

## Additional Resources

- [EXECUTION.md](EXECUTION.md) - Subagent prompt templates
- [ARTIFACTS.md](ARTIFACTS.md) - Artifact management details
- [accelyst-guide](../accelyst-guide/SKILL.md) - CLI operations and architecture
