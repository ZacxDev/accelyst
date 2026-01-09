---
name: milestone-convert
description: Convert feature backlogs, TODO lists, or scratch notes into Accelyst milestone YAML format. Use when the user wants to transform unstructured feature ideas into structured milestones, or when they mention "convert to milestone", "accelyst format", or "milestone yaml".
allowed-tools: Read, Write, Glob
---

# Milestone Format Converter

Convert unstructured feature backlogs into the Accelyst epic YAML format.

## Architecture Overview

Accelyst uses a registry + epics architecture:

- **Registry** (`.accelyst/registry.yaml`) - Persistent project knowledge
  - Project parts: logical components (client, api, database, etc.)
  - Artifacts: analysis outputs that persist across sessions
- **Epics** (`.accelyst/epics/*.yaml`) - Ephemeral milestone collections
  - Each epic file contains related milestones
  - Epics are disposable and regeneratable
- **Artifacts** (`.accelyst/artifacts/*.md`) - Knowledge outputs
  - Steps can `produces` artifacts (saved during execution)
  - Steps can `requires` artifacts (injected into prompt context)

## Process

1. **Read the input** - Accept feature lists, TODO files, scratch notes, or any unstructured feature description
2. **Check existing registry** - Read `.accelyst/registry.yaml` for existing project parts and artifacts
3. **Identify new project parts** - Extract logical components not already in registry
4. **Group into milestones** - Cluster related features into coherent milestones
5. **Identify milestone dependencies** - Determine which milestones must complete before others can start
6. **Decompose into steps** - Break each milestone into atomic, actionable steps
7. **Add analysis steps with produces** - Analysis steps should produce artifacts for future reference
8. **Add requires for dependent steps** - Steps that need prior analysis should require those artifacts
9. **Add test steps** - Every milestone must have at least one step referencing the `test` project part
10. **Map step dependencies** - Identify which steps depend on others within each milestone
11. **Assign project parts** - Link each step to relevant project components
12. **Generate output** - Output epic YAML and registry updates (if needed)

## Output Files

### Epic File (`.accelyst/epics/<epic-name>.yaml`)

```yaml
name: "<Epic Name>"
description: "<Brief description of this epic>"
priority: ready  # ready | in_progress | blocked | ideas

milestones:
  - id: <milestone_id>
    name: "<Feature Name>"
    dependsOn: []  # or list of milestone IDs within this epic
    steps:
      - id: <step_id>
        instruction: "<action to perform>"
        dependsOn: []
        projectParts:
          - <component-name>
        produces: <artifact_id>  # optional: artifact this step creates
        requires:                # optional: artifacts this step needs
          - <artifact_id>
```

### Registry Updates (if new project parts needed)

If new project parts are discovered, output the additions for `.accelyst/registry.yaml`:

```yaml
# Add to .accelyst/registry.yaml projectParts:
  - name: <new-component>
    directoryAbs: <absolute-path>
    description: "<component description>"
```

## Milestone Organization

Milestones form a DAG (directed acyclic graph) for execution ordering:

- **Independent milestones** (`dependsOn: []`) can execute in parallel
- **Dependent milestones** wait for their dependencies to complete
- Group related work into milestones that can be reasoned about as a unit

## Step Decomposition Guidelines

For each milestone, create steps following this pattern:

1. **Analysis steps** (no dependencies) - Research, documentation lookup, codebase analysis
   - These should `produces` artifacts capturing insights
2. **Design steps** (depend on analysis) - Architecture decisions, schema design
   - These should `requires` analysis artifacts
3. **Implementation steps** (depend on design) - Core feature implementation
4. **Testing steps** (depend on implementation) - Unit tests, integration tests, E2E tests (REQUIRED)
5. **Integration steps** (depend on testing) - Connecting components, API wiring
6. **Polish steps** (depend on integration) - UI refinement, error handling, edge cases

**Testing is mandatory**: Every milestone must include at least one step that references the `test` project part.

## Artifact System

### produces Field

When a step generates reusable knowledge (analysis, architecture decisions, schemas):

```yaml
- id: analyze_auth
  instruction: "analyze the existing authentication implementation"
  produces: auth_analysis
  projectParts:
    - api
```

The artifact is saved to `.accelyst/artifacts/auth_analysis.md` and registered.

### requires Field

When a step needs prior analysis:

```yaml
- id: impl_oauth
  instruction: "implement OAuth2 integration"
  requires:
    - auth_analysis
  dependsOn:
    - analyze_auth
  projectParts:
    - api
```

The artifact content is injected into the prompt context.

## Dependency Rules

**Milestone-level:**
- Milestones with no dependencies can run in parallel
- Use milestone dependencies for cross-cutting concerns (e.g., "UI" depends on "API")
- Never create circular dependencies between milestones

**Step-level (within a milestone):**
- Steps with no dependencies can run in parallel (marked for subagent execution)
- Analysis steps typically have no dependencies
- Implementation steps depend on relevant analysis steps
- Step dependencies are scoped to the containing milestone only

## Example Transformation

**Input:**
```
- add dark mode toggle to settings
- need to update CSS variables
- persist preference in localStorage
- also need API endpoint to sync preference across devices
```

**Output Epic** (`.accelyst/epics/dark-mode.yaml`):

```yaml
name: "Dark Mode Feature"
description: "Add dark mode toggle with cross-device sync"
priority: ready

milestones:
  - id: dark_mode_api
    name: "Dark Mode API"
    dependsOn: []
    steps:
      - id: analyze_preferences
        instruction: "analyze existing user preferences storage and API patterns"
        produces: preferences_analysis
        dependsOn: []
        projectParts:
          - api
      - id: design_schema
        instruction: "design user preferences schema for theme storage"
        requires:
          - preferences_analysis
        dependsOn:
          - analyze_preferences
        projectParts:
          - api
      - id: impl_endpoint
        instruction: "implement preference sync endpoint"
        dependsOn:
          - design_schema
        projectParts:
          - api
      - id: test_endpoint
        instruction: "write unit and integration tests for preference sync endpoint"
        dependsOn:
          - impl_endpoint
        projectParts:
          - test

  - id: dark_mode_ui
    name: "Dark Mode Toggle"
    dependsOn: []
    steps:
      - id: analyze_theme
        instruction: "analyze current theme and styling implementation"
        produces: theme_analysis
        dependsOn: []
        projectParts:
          - client
      - id: define_css_vars
        instruction: "define CSS custom properties for light and dark themes"
        requires:
          - theme_analysis
        dependsOn:
          - analyze_theme
        projectParts:
          - client
      - id: impl_toggle
        instruction: "implement theme toggle component in settings page"
        dependsOn:
          - define_css_vars
        projectParts:
          - client
      - id: add_persistence
        instruction: "add localStorage persistence for theme preference"
        dependsOn:
          - impl_toggle
        projectParts:
          - client
      - id: test_toggle
        instruction: "write component tests for theme toggle and persistence"
        dependsOn:
          - add_persistence
        projectParts:
          - test

  - id: dark_mode_sync
    name: "Dark Mode Sync"
    dependsOn:
      - dark_mode_api
      - dark_mode_ui
    steps:
      - id: impl_sync
        instruction: "integrate preference sync with API endpoint"
        requires:
          - preferences_analysis
          - theme_analysis
        dependsOn: []
        projectParts:
          - client
      - id: handle_conflicts
        instruction: "handle conflicts between local and server preferences"
        dependsOn:
          - impl_sync
        projectParts:
          - client
      - id: test_sync
        instruction: "write integration tests for preference sync including conflict scenarios"
        dependsOn:
          - handle_conflicts
        projectParts:
          - test
```

## Quality Checklist

Before outputting, verify:

- [ ] All milestone IDs are unique within the epic
- [ ] All milestone dependsOn reference valid milestone IDs in the same epic
- [ ] All step IDs are unique within their milestone
- [ ] All step dependsOn reference valid step IDs in the same milestone
- [ ] All projectParts reference parts defined in the registry
- [ ] No circular dependencies exist (milestone or step level)
- [ ] Independent milestones have `dependsOn: []` (parallelizable)
- [ ] Analysis/research steps have no dependencies (parallelizable)
- [ ] Analysis steps use `produces` for reusable insights
- [ ] Implementation steps use `requires` when they need prior analysis
- [ ] Implementation steps depend on their prerequisites
- [ ] Instructions are actionable and specific
- [ ] **`test` project part exists in registry** (REQUIRED)
- [ ] **Every milestone has at least one step referencing `test`** (REQUIRED)

For the complete specification, see [SPEC.md](../../../SPEC.md).

## Related Skills

- `/milestone-orchestrate` - **Full-cycle**: runs plan → convert → execute automatically
- `/milestone-plan` - Interactive backlog discovery → planning brief (previous step)
- `/milestone-execute` - Execute epic via subagents (next step)
- `/accelyst-guide` - CLI operations and troubleshooting
