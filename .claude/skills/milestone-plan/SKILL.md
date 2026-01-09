---
name: milestone-plan
description: Interactive discovery skill that transforms messy backlogs, scratch notes, or feature ideas into structured planning briefs. Use when input needs clarification before conversion to Accelyst format, or when the user has "needs planning" items. Outputs a clean brief for milestone-convert.
allowed-tools: Read, Write, Glob, AskUserQuestion
---

# Milestone Planning Assistant

Transform unstructured feature backlogs into structured planning briefs through interactive discovery.

## When to Use

- Input contains mixed maturity levels (ready vs needs planning vs ideas)
- Input has ambiguous scope or granularity
- Project parts need to be identified from context
- Dependencies between items are implicit
- User wants guidance on how to structure their backlog

## Process

1. **Read & Parse Input** - Analyze the input file/text structure
2. **Detect Format** - Identify sections, parts, bullets, nesting patterns
3. **Extract Project Parts** - Scan for service/system references
4. **Identify Ambiguities** - Flag items needing clarification
5. **Ask Questions** - Use AskUserQuestion for critical decisions
6. **Output Brief** - Generate structured planning brief

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
"I found references to these systems: [api, client, dspy-service, semantic-service].
What are their directory paths? Are there others I missed?"
```

### Section Filtering
```
"Your input has sections: 'ready', 'needs planning', 'ideas'.
Which sections should I include in this planning session?"
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

projectParts:
  - name: client
    directoryAbs: /path/to/client
    description: "Frontend React application"
  - name: api
    directoryAbs: /path/to/api
    description: "Go backend API"
  - name: test
    directoryAbs: /path/to/tests
    description: "Test suites"

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
        projectParts: [client]
      - id: impl_download
        instruction: "implement download button in fullscreen view"
        dependsOn: [analyze_fullscreen]
        projectParts: [client]
      - id: impl_zoom
        instruction: "implement pinch/scroll zoom with viewport controls"
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
        projectParts: [client]
      - id: impl_syntax_highlighting
        instruction: "implement sd-prompt syntax highlighting with dataset token colors"
        dependsOn: [analyze_codemirror]
        projectParts: [client]
      # ... more steps
    dependsOn: []
    notes: "Large feature - consider breaking into phases"
```

## Interaction Flow

### Step 1: Initial Analysis
```
I've analyzed your input file. Here's what I found:

**Sections:**
- ready (15 items)
- needs planning (22 items)
- ideas (18 items)

**Detected Project Parts:**
- client (referenced 28 times)
- api (referenced 12 times)
- dspy-service (referenced 8 times)
- semantic-service (referenced 5 times)

**Groupings detected:**
- fullscreen view (4 items)
- model selection (6 items)
- index page (8 items)
- CodeMirror Prompt IDE (12 items)
```

### Step 2: Ask Configuration Questions
Use AskUserQuestion to get:
1. Which sections to include
2. Project part paths
3. Grouping preferences for large feature sets

### Step 3: Ask Granularity Questions
For each ambiguous grouping, ask how to structure it.

### Step 4: Generate Brief
Output the planning brief YAML to a file (e.g., `planning-brief.yaml`).

### Step 5: Suggest Next Steps
```
Planning brief saved to: planning-brief.yaml

Next steps:
1. Review the brief and adjust as needed
2. Run `/milestone-convert planning-brief.yaml` to generate Accelyst YAML
```

## Handling "Needs Planning" Items

Items marked as "needs planning" require extra care:

1. **Flag them explicitly** - Include `priority: needs_planning` in output
2. **Add notes** - Capture what planning is needed
3. **Suggest research steps** - Add analysis/research steps before implementation
4. **Ask about deferral** - "Should 'needs planning' items be deferred to a future session?"

## Handling Completion Markers

When `(done)` markers are found:

1. **Filter by default** - Don't include completed items in milestones
2. **Track context** - Note that partial completion exists
3. **Ask if relevant** - "Some items are marked done. Include them for reference?"

## Quality Checks

Before outputting, verify:

- [ ] All detected project parts have paths (or are marked TBD)
- [ ] Large groupings have been addressed (grouped or split)
- [ ] Dependencies between milestones are identified
- [ ] "Needs planning" items are flagged appropriately
- [ ] Test project part is included
- [ ] Each milestone has suggested test steps

## Example Session

**User:** `/milestone-plan /path/to/scratch.txt`

**Assistant:**
```
I've analyzed scratch.txt. Found 3 sections with 55 total items.

Let me ask a few questions to create your planning brief:
```

[AskUserQuestion: Which sections to include?]
[AskUserQuestion: Project part paths?]
[AskUserQuestion: How to handle 'CodeMirror Prompt IDE' (12 items)?]

```
Based on your answers, I've created a planning brief with:
- 8 milestones from 'ready' section
- 6 milestones from 'needs planning' section (flagged for review)
- Skipped 'ideas' section as requested

Brief saved to: planning-brief.yaml

To generate Accelyst YAML: /milestone-convert planning-brief.yaml
```

## Integration with milestone-convert

The planning brief format is designed to be directly consumable by `milestone-convert`:

1. `projectParts` maps directly to Accelyst format
2. `suggestedSteps` become the milestone steps
3. `dependsOn` relationships are preserved
4. `notes` are dropped (they're for human review only)

The user can:
- Edit the brief before conversion
- Run convert directly on the brief
- Iterate on specific milestones
