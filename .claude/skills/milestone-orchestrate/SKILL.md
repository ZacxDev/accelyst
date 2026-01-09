---
name: milestone-orchestrate
description: Full-cycle feature implementation from raw backlog to working code. Orchestrates the complete skill chain (plan → convert → execute) to transform unstructured feature ideas into implemented features. Use when the user has a scratch file, backlog, or feature list they want fully implemented.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, Task, TodoWrite, AskUserQuestion, Skill
---

# Milestone Orchestrator

Transform raw feature backlogs into implemented code through the complete Accelyst skill chain.

## Overview

This skill orchestrates the full implementation lifecycle:

```
┌─────────────────────────────────────────────────────────────────┐
│                    /milestone-orchestrate                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Raw Backlog ──► /milestone-plan ──► Planning Brief            │
│        │                                    │                    │
│        │         (interactive discovery)    │                    │
│        │                                    ▼                    │
│        │              /milestone-convert ◄──┘                    │
│        │                     │                                   │
│        │         (YAML generation)                               │
│        │                     ▼                                   │
│        │              Epic YAML file                             │
│        │                     │                                   │
│        │              /milestone-execute                         │
│        │                     │                                   │
│        │         (subagent orchestration)                        │
│        │                     ▼                                   │
│        └──────────► Implemented Features                         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Invocation

```bash
# Full orchestration from scratch file
/milestone-orchestrate <path-to-backlog>

# Example
/milestone-orchestrate scratch.txt
/milestone-orchestrate features.md
/milestone-orchestrate planning-notes.yaml

# With options
/milestone-orchestrate scratch.txt --epic-name my-feature
/milestone-orchestrate scratch.txt --skip-plan  # If planning brief already exists
/milestone-orchestrate scratch.txt --skip-convert  # If epic YAML already exists
```

## Process

### Phase 1: Initialization

1. **Validate input file exists**
   ```
   Read <input-file>
   → Confirm file is readable
   → Detect format (txt, md, yaml)
   ```

2. **Check Accelyst project state**
   ```bash
   # Ensure project is initialized
   ls .accelyst/registry.yaml || ./accelyst init

   # Check existing epics
   ./accelyst epics list
   ```

3. **Create orchestration plan**
   ```
   TodoWrite:
   - [ ] Phase 2: Run milestone-plan (backlog → brief)
   - [ ] Phase 3: Run milestone-convert (brief → epic YAML)
   - [ ] Phase 4: Run milestone-execute (epic → implementation)
   - [ ] Phase 5: Verify and summarize
   ```

### Phase 2: Planning (milestone-plan)

**Goal:** Transform raw backlog into structured planning brief

```
Skill(
  skill: "milestone-plan",
  args: "<input-file>"
)
```

**Expected outputs:**
- Planning brief YAML file (e.g., `planning-brief.yaml`)
- Registry updates (new project parts identified)
- Questions answered via AskUserQuestion

**Checkpoint:** Verify planning brief was created before proceeding.

### Phase 3: Conversion (milestone-convert)

**Goal:** Convert planning brief to epic YAML

```
Skill(
  skill: "milestone-convert",
  args: "<planning-brief-path>"
)
```

**Expected outputs:**
- Epic YAML file at `.accelyst/epics/<name>.yaml`
- Updated registry if new project parts added

**Checkpoint:** Verify epic file exists and validates:
```bash
./accelyst epics list
```

### Phase 4: Execution (milestone-execute)

**Goal:** Execute the epic via subagents

```
Skill(
  skill: "milestone-execute",
  args: "<epic-name>"
)
```

**Expected outputs:**
- Implemented features
- Artifacts in `.accelyst/artifacts/`
- Updated registry with produced artifacts
- Passing tests (if test steps included)

**Checkpoint:** All TodoWrite items from execution marked complete.

### Phase 5: Verification

1. **Verify all artifacts produced**
   ```bash
   ./accelyst registry artifacts
   ls .accelyst/artifacts/
   ```

2. **Run project validation** (if applicable)
   ```bash
   # Language-specific validation
   go build ./...        # Go
   npm run build         # Node.js
   python -m py_compile  # Python
   ```

3. **Generate summary report**

## Interaction Points

The orchestrator pauses for user input at key decision points:

### After Planning (Phase 2)

```
Planning complete. Generated brief: planning-brief.yaml

Milestones identified:
1. auth_api (5 steps)
2. auth_ui (4 steps)
3. auth_tests (3 steps)

Proceed with conversion? [Yes / Review brief first / Modify / Abort]
```

### After Conversion (Phase 3)

```
Epic created: .accelyst/epics/auth-feature.yaml

Milestones ready for execution:
1. auth_api (ready, no dependencies)
2. auth_ui (depends on: auth_api)
3. auth_tests (depends on: auth_api, auth_ui)

Execute which milestones?
1. All milestones sequentially (Recommended)
2. Specific milestone only
3. Review epic first
4. Abort
```

### During Execution (Phase 4)

For each milestone, show progress and allow intervention:
```
Milestone 1/3: auth_api

[✓] analyze_auth - Complete (artifact: auth_analysis)
[→] implement_endpoints - In progress...
[ ] test_auth - Pending

Continue? [Yes / Pause / Skip to next milestone / Abort]
```

## Error Recovery

### Planning Fails

```
Milestone-plan encountered an error:
{error_details}

Options:
1. Retry planning with different approach
2. Create planning brief manually
3. Skip planning (use existing brief)
4. Abort orchestration
```

### Conversion Fails

```
Milestone-convert encountered an error:
{error_details}

Options:
1. Retry conversion
2. Edit planning brief and retry
3. Create epic YAML manually
4. Abort orchestration
```

### Execution Fails

```
Milestone-execute failed at step: {step_id}
{error_details}

Options:
1. Retry this step
2. Skip step and continue
3. Mark milestone as partial and continue to next
4. Abort orchestration
```

## Skip Modes

### --skip-plan

Use when planning brief already exists:
```bash
/milestone-orchestrate scratch.txt --skip-plan --brief planning-brief.yaml
```

Starts from Phase 3 (conversion).

### --skip-convert

Use when epic YAML already exists:
```bash
/milestone-orchestrate scratch.txt --skip-convert --epic auth-feature
```

Starts from Phase 4 (execution).

### --plan-only

Stop after planning:
```bash
/milestone-orchestrate scratch.txt --plan-only
```

Outputs planning brief without conversion or execution.

### --convert-only

Stop after conversion (requires --skip-plan or existing brief):
```bash
/milestone-orchestrate --skip-plan --brief planning-brief.yaml --convert-only
```

## Example Session

**User:** `/milestone-orchestrate scratch.txt`

```
=== Milestone Orchestrator ===

Phase 1: Initialization
✓ Input file: scratch.txt (readable)
✓ Accelyst project: initialized
✓ Registry: 5 project parts, 2 artifacts

Creating orchestration plan...

[TodoWrite: 5 phases created]

─────────────────────────────────────────

Phase 2: Planning

Running /milestone-plan on scratch.txt...

[Skill: milestone-plan invoked]

Planning complete!
- Brief saved to: planning-brief.yaml
- New project parts identified: 2 (semantic_service, dspy_service)
- Milestones planned: 4

Proceed to conversion? [Yes]

─────────────────────────────────────────

Phase 3: Conversion

Running /milestone-convert on planning-brief.yaml...

[Skill: milestone-convert invoked]

Conversion complete!
- Epic saved to: .accelyst/epics/feature-enhancements.yaml
- Milestones: 4
- Total steps: 18
- Artifacts to produce: 5

Proceed to execution? [Yes]

─────────────────────────────────────────

Phase 4: Execution

Running /milestone-execute on feature-enhancements...

[Skill: milestone-execute invoked]

Executing milestone 1/4: fullscreen_download
  [✓] analyze_fullscreen
  [✓] impl_download_btn
  [✓] test_download

Executing milestone 2/4: model_selection_ux
  [✓] analyze_model_selector
  [✓] impl_portrait_placeholders
  [✓] test_portrait_placeholders

Executing milestone 3/4: image_grid_performance
  [✓] analyze_grid
  [✓] impl_append_last_run
  [✓] impl_viewport_rendering
  [✓] test_grid_performance

Executing milestone 4/4: tools_page_rename
  [✓] analyze_prompt_doctor
  [✓] rename_route
  [✓] update_copy
  [✓] test_rename

─────────────────────────────────────────

Phase 5: Verification

Artifacts produced: 4
- fullscreen_view_analysis
- model_selector_analysis
- image_grid_analysis
- prompt_doctor_analysis

Registry updated: Yes

Build validation: ✓ Passed

─────────────────────────────────────────

=== Orchestration Complete ===

Summary:
- Input: scratch.txt
- Milestones executed: 4/4
- Steps completed: 18/18
- Artifacts produced: 4
- Tests passing: Yes

Files created/modified:
- .accelyst/epics/feature-enhancements.yaml
- .accelyst/artifacts/*.md (4 files)
- client/*.js (implementation files)
- client/e2e/*.spec.ts (test files)
```

## Quality Checks

**Before starting:**
- [ ] Input file exists and is readable
- [ ] Accelyst project initialized
- [ ] Registry has at least one project part

**After each phase:**
- [ ] Phase output file exists
- [ ] No critical errors encountered
- [ ] User confirmed to proceed (interactive mode)

**Final verification:**
- [ ] All planned milestones executed
- [ ] All expected artifacts exist
- [ ] Registry updated with new artifacts
- [ ] Build/validation passes (if applicable)

## Integration

### Skill Dependencies

| Skill | Purpose | Invoked Via |
|-------|---------|-------------|
| `/accelyst-guide` | CLI operations | Direct commands |
| `/milestone-plan` | Backlog → Brief | `Skill` tool |
| `/milestone-convert` | Brief → Epic | `Skill` tool |
| `/milestone-execute` | Epic → Code | `Skill` tool |

### Workflow Variations

**Full automation:**
```
/milestone-orchestrate backlog.txt --auto
```
Runs all phases without pausing (except for critical errors).

**Incremental development:**
```
/milestone-orchestrate backlog.txt --plan-only
# Review and edit planning-brief.yaml
/milestone-orchestrate --skip-plan --brief planning-brief.yaml --convert-only
# Review and edit epic YAML
/milestone-orchestrate --skip-plan --skip-convert --epic my-feature
```

**Resume after failure:**
```
# If execution failed at milestone 3
/milestone-orchestrate --skip-plan --skip-convert --epic my-feature --start-at milestone_3
```

## Related Skills

- `/accelyst-guide` - CLI operations and troubleshooting
- `/milestone-plan` - Interactive planning (Phase 2)
- `/milestone-convert` - YAML conversion (Phase 3)
- `/milestone-execute` - Subagent execution (Phase 4)
