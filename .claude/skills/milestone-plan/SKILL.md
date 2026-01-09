---
name: milestone-plan
description: Interactive discovery skill that transforms messy backlogs, scratch notes, or feature ideas into structured planning briefs. Use when input needs clarification before conversion to Accelyst format, or when the user has "needs planning" items. Outputs a clean brief for milestone-convert.
allowed-tools: Read, Write, Glob, AskUserQuestion
---

# Milestone Planning Assistant

Transform unstructured feature backlogs into structured planning briefs through interactive discovery.

## Architecture Overview

Accelyst uses a registry + epics architecture:

- **Registry** (`.accelyst/registry.yaml`) - Persistent project knowledge
  - Project parts: logical components (client, api, database, etc.)
  - Artifacts: analysis outputs that persist across sessions
- **Epics** (`.accelyst/epics/*.yaml`) - Ephemeral milestone collections
  - Each epic file contains related milestones
  - Multiple epics can exist in parallel
- **Artifacts** (`.accelyst/artifacts/*.md`) - Knowledge outputs
  - Steps can `produces` artifacts (saved during execution)
  - Steps can `requires` artifacts (injected into prompt context)

## When to Use

- Input contains mixed maturity levels (ready vs needs planning vs ideas)
- Input has ambiguous scope or granularity
- Project parts need to be identified from context
- Dependencies between items are implicit
- User wants guidance on how to structure their backlog
- Existing registry needs to be checked for project parts/artifacts

## Process

1. **Check Registry** - Read `.accelyst/registry.yaml` for existing project parts and artifacts
2. **Read & Parse Input** - Analyze the input file/text structure
3. **Detect Format** - Identify sections, parts, bullets, nesting patterns
4. **Extract Project Parts** - Scan for service/system references, compare with registry
5. **Identify Artifacts** - Check if any analysis artifacts already exist that could be reused
6. **Identify Ambiguities** - Flag items needing clarification
7. **Ask Questions** - Use AskUserQuestion for critical decisions
8. **Output Brief** - Generate structured planning brief

## Input Format Detection

The skill recognizes common backlog patterns:

```
## section          → Maturity/priority level (ready, needs planning, ideas)
*part*              → Logical grouping (UI area, service, feature set)
- item              → Individual feature/task
  - nested item     → Sub-task, detail, or acceptance criteria
(done)              → Completion marker (skip or filter)
```

## Questions to Ask

### Project Parts Discovery
```
"Existing registry has: [api, client, database].
I found references to: [semantic-service, dspy-service].
Should I add these to the registry? What are their directory paths?"
```

### Artifact Reuse
```
"Existing artifacts: [api_architecture, client_state_analysis].
These may be relevant to your new features. Should any steps require them?"
```

### Section Filtering
```
"Your input has sections: 'ready', 'needs planning', 'ideas'.
Which sections should I include in this epic?"
```

### Epic Organization
```
"Should this become:
1. A single epic with all features
2. Multiple epics grouped by theme (e.g., 'ui-enhancements', 'api-features')
3. A new epic that adds to existing epics"
```

### Granularity Decisions
```
"'CodeMirror Prompt IDE' has 12 nested items. Should this be:
1. One milestone with 12 steps
2. Multiple milestones grouped by sub-feature
3. Split into separate milestones for each major capability"
```

### Dependency Clarification
```
"'optimize prompt checkbox' references 'prompt codemirror'.
Does this depend on the CodeMirror implementation being complete?"
```

### Grouping Strategy
```
"These 5 items all relate to 'fullscreen view':
- download button
- upscale section image viewer
- pinch/scroll zoom
- select two images for video
Should they be one milestone or separate?"
```

## Output: Planning Brief

Generate a structured YAML brief that `milestone-convert` can consume directly:

```yaml
# Planning Brief for: [Project Name]
# Generated: [timestamp]
# Sections included: [ready, needs planning]
# Epic name: [suggested-epic-name]

# Registry updates needed (if any new project parts):
registryUpdates:
  projectParts:
    - name: semantic-service
      directoryAbs: /path/to/semantic-service
      description: "Semantic search and embedding service"

# Existing artifacts that may be useful:
availableArtifacts:
  - id: api_architecture
    description: "Analysis of API structure and patterns"
  - id: client_state_analysis
    description: "Frontend state management analysis"

milestones:
  - id: fullscreen_view_enhancements
    name: "Fullscreen View Enhancements"
    description: "Improve the fullscreen image viewing experience"
    priority: ready
    items:
      - "download button"
      - "upscale section: clicking image opens dedicated page"
      - "pinch/scroll zoom with tap/drag viewport control"
    suggestedSteps:
      - id: analyze_fullscreen
        instruction: "analyze current fullscreen view implementation"
        produces: fullscreen_analysis  # Creates new artifact
        projectParts: [client]
      - id: impl_download
        instruction: "implement download button in fullscreen view"
        requires: [fullscreen_analysis]  # Uses the artifact
        dependsOn: [analyze_fullscreen]
        projectParts: [client]
      - id: impl_zoom
        instruction: "implement pinch/scroll zoom with viewport controls"
        requires: [fullscreen_analysis]
        dependsOn: [analyze_fullscreen]
        projectParts: [client]
      - id: test_fullscreen
        instruction: "write tests for fullscreen view enhancements"
        dependsOn: [impl_download, impl_zoom]
        projectParts: [test]
    dependsOn: []
    notes: "User chose to group all fullscreen items together"

  - id: codemirror_prompt_ide
    name: "CodeMirror Prompt IDE"
    description: "Advanced prompt editing with syntax highlighting and autocomplete"
    priority: needs_planning
    items:
      - "sd-prompt syntax highlighting"
      - "SDXL weight syntax highlighting"
      - "prompt writer copilot"
      - "semantic search for token combinations"
      - "warning underlines for format issues"
    suggestedSteps:
      - id: analyze_codemirror
        instruction: "analyze current CodeMirror integration and prompt textarea"
        produces: codemirror_analysis
        projectParts: [client]
      - id: impl_syntax_highlighting
        instruction: "implement sd-prompt syntax highlighting with dataset token colors"
        requires:
          - codemirror_analysis
          - client_state_analysis  # Existing artifact
        dependsOn: [analyze_codemirror]
        projectParts: [client]
      # ... more steps
    dependsOn: []
    notes: "Large feature - consider breaking into phases"
```

## Interaction Flow

### Step 1: Check Existing Registry
```
I checked .accelyst/registry.yaml:

**Existing Project Parts:**
- client: /path/to/client
- api: /path/to/api
- test: /path/to/tests

**Existing Artifacts:**
- api_architecture: API structure analysis (from auth_milestone/analyze_api)
- database_schema: Schema documentation (from data_milestone/analyze_db)
```

### Step 2: Initial Analysis
```
I've analyzed your input file. Here's what I found:

**Sections:**
- ready (15 items)
- needs planning (22 items)
- ideas (18 items)

**New Project Parts Detected:**
- semantic-service (referenced 5 times) - not in registry
- dspy-service (referenced 8 times) - not in registry

**Groupings detected:**
- fullscreen view (4 items)
- model selection (6 items)
- index page (8 items)
- CodeMirror Prompt IDE (12 items)

**Potentially Useful Artifacts:**
- api_architecture could help with API integration steps
```

### Step 3: Ask Configuration Questions
Use AskUserQuestion to get:
1. Which sections to include
2. New project part paths (for registry updates)
3. Which existing artifacts to require
4. Epic name and organization
5. Grouping preferences for large feature sets

### Step 4: Ask Granularity Questions
For each ambiguous grouping, ask how to structure it.

### Step 5: Generate Brief
Output the planning brief YAML to a file (e.g., `planning-brief.yaml`).

### Step 6: Suggest Next Steps
```
Planning brief saved to: planning-brief.yaml

Next steps:
1. Review the brief and adjust as needed
2. Run `/milestone-convert planning-brief.yaml` to generate epic YAML
3. Add any new project parts to .accelyst/registry.yaml
```

## Handling "Needs Planning" Items

Items marked as "needs planning" require extra care:

1. **Flag them explicitly** - Include `priority: needs_planning` in output
2. **Add notes** - Capture what planning is needed
3. **Suggest research steps** - Add analysis/research steps with `produces` before implementation
4. **Ask about deferral** - "Should 'needs planning' items be deferred to a separate epic?"

## Handling Completion Markers

When `(done)` markers are found:

1. **Filter by default** - Don't include completed items in milestones
2. **Track context** - Note that partial completion exists
3. **Ask if relevant** - "Some items are marked done. Include them for reference?"

## Artifact Strategy

When creating the brief:

1. **Analysis steps should produce** - Every analysis step should create an artifact
2. **Implementation can require** - Steps that depend on analysis should require the artifact
3. **Cross-milestone sharing** - Artifacts from one milestone can be required by another
4. **Registry artifacts** - Existing artifacts in the registry can be required immediately

## Quality Checks

Before outputting, verify:

- [ ] Registry was checked for existing project parts
- [ ] Registry was checked for existing artifacts
- [ ] All detected project parts have paths (or are marked as registry updates)
- [ ] Analysis steps use `produces` for reusable insights
- [ ] Implementation steps use `requires` where applicable
- [ ] Large groupings have been addressed (grouped or split)
- [ ] Dependencies between milestones are identified
- [ ] "Needs planning" items are flagged appropriately
- [ ] Test project part is included
- [ ] Each milestone has suggested test steps

## Example Session

**User:** `/milestone-plan /path/to/scratch.txt`

**Assistant:**
```
I checked .accelyst/registry.yaml:
- Found 3 project parts: client, api, test
- Found 2 artifacts: api_architecture, auth_patterns

Analyzing scratch.txt... Found 3 sections with 55 total items.

Let me ask a few questions to create your planning brief:
```

[AskUserQuestion: Which sections to include?]
[AskUserQuestion: New project part paths?]
[AskUserQuestion: Should 'api_architecture' artifact be reused?]
[AskUserQuestion: How to handle 'CodeMirror Prompt IDE' (12 items)?]
[AskUserQuestion: Epic name?]

```
Based on your answers, I've created a planning brief with:
- Epic name: "ui-enhancements"
- 8 milestones from 'ready' section
- 6 milestones from 'needs planning' section (flagged for review)
- Skipped 'ideas' section as requested
- 3 new project parts to add to registry
- 4 steps will require existing api_architecture artifact

Brief saved to: planning-brief.yaml

To generate epic YAML: /milestone-convert planning-brief.yaml
```

## Integration with milestone-convert

The planning brief format is designed to be directly consumable by `milestone-convert`:

1. `registryUpdates` lists new project parts to add to registry
2. `availableArtifacts` documents artifacts available for `requires`
3. `suggestedSteps` become the milestone steps
4. `produces`/`requires` relationships are preserved
5. `dependsOn` relationships are preserved
6. `notes` are dropped (they're for human review only)

The user can:
- Edit the brief before conversion
- Run convert directly on the brief
- Iterate on specific milestones
