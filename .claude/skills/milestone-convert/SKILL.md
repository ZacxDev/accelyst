---
name: milestone-convert
description: Convert feature backlogs, TODO lists, or scratch notes into Accelyst milestone YAML format. Use when the user wants to transform unstructured feature ideas into structured milestones, or when they mention "convert to milestone", "accelyst format", or "milestone yaml".
allowed-tools: Read, Write, Glob
---

# Milestone Format Converter

Convert unstructured feature backlogs into the Accelyst milestone YAML format.

## Process

1. **Read the input** - Accept feature lists, TODO files, scratch notes, or any unstructured feature description
2. **Identify project parts** - Extract logical components (client, api, database, services, etc.) + always include `test`
3. **Group into milestones** - Cluster related features into coherent milestones
4. **Identify milestone dependencies** - Determine which milestones must complete before others can start
5. **Decompose into steps** - Break each milestone into atomic, actionable steps
6. **Add test steps** - Every milestone must have at least one step referencing the `test` project part
7. **Map step dependencies** - Identify which steps depend on others within each milestone
8. **Assign project parts** - Link each step to relevant project components
9. **Generate YAML** - Output valid Accelyst configuration

## Milestone Organization

Milestones form a DAG (directed acyclic graph) for execution ordering:

- **Independent milestones** (`dependsOn: []`) can execute in parallel
- **Dependent milestones** wait for their dependencies to complete
- Group related work into milestones that can be reasoned about as a unit

## Step Decomposition Guidelines

For each milestone, create steps following this pattern:

1. **Analysis steps** (no dependencies) - Research, documentation lookup, codebase analysis
2. **Design steps** (depend on analysis) - Architecture decisions, schema design
3. **Implementation steps** (depend on design) - Core feature implementation
4. **Testing steps** (depend on implementation) - Unit tests, integration tests, E2E tests (REQUIRED)
5. **Integration steps** (depend on testing) - Connecting components, API wiring
6. **Polish steps** (depend on integration) - UI refinement, error handling, edge cases

**Testing is mandatory**: Every milestone must include at least one step that references the `test` project part. This ensures all features have test coverage by default.

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

## Output Format

```yaml
projectParts:
  - name: <component-name>
    directoryAbs: <absolute-path>
    documentationAbs: <docs-path>  # optional
  - name: test                     # REQUIRED: always include test project part
    directoryAbs: <path-to-tests>

milestones:
  - id: <milestone_id>
    name: "<Feature Name>"
    dependsOn: []  # or list of milestone IDs
    steps:
      - id: <step_id>
        instruction: "<action to perform>"
        dependsOn: []
        projectParts:
          - <component-name>
      - id: <test_step_id>         # REQUIRED: at least one test step per milestone
        instruction: "<write tests for ...>"
        dependsOn:
          - <implementation_step>
        projectParts:
          - test
```

## Example Transformation

**Input:**
```
- add dark mode toggle to settings
- need to update CSS variables
- persist preference in localStorage
- also need API endpoint to sync preference across devices
```

**Output:**
```yaml
projectParts:
  - name: client
    directoryAbs: /path/to/client
  - name: api
    directoryAbs: /path/to/api
  - name: test
    directoryAbs: /path/to/tests

milestones:
  - id: dark_mode_api
    name: "Dark Mode API"
    dependsOn: []
    steps:
      - id: design_schema
        instruction: "design user preferences schema"
        dependsOn: []
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
        dependsOn: []
        projectParts:
          - client
      - id: define_css_vars
        instruction: "define CSS custom properties for light and dark themes"
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

- [ ] All milestone IDs are unique across the configuration
- [ ] All milestone dependsOn reference valid milestone IDs
- [ ] All step IDs are unique within their milestone
- [ ] All step dependsOn reference valid step IDs in the same milestone
- [ ] All projectParts reference defined project parts
- [ ] No circular dependencies exist (milestone or step level)
- [ ] Independent milestones have `dependsOn: []` (parallelizable)
- [ ] Analysis/research steps have no dependencies (parallelizable)
- [ ] Implementation steps depend on their prerequisites
- [ ] Instructions are actionable and specific
- [ ] **`test` project part is defined** (REQUIRED)
- [ ] **Every milestone has at least one step referencing `test`** (REQUIRED)

For the complete specification, see [SPEC.md](../../../SPEC.md).
