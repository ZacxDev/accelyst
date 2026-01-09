---
name: milestone-plan
description: Interactive discovery skill that transforms messy backlogs, scratch notes, or feature ideas into structured planning briefs. Use when input needs clarification before conversion to Accelyst format, or when the user has "needs planning" items. Outputs a clean brief for milestone-convert.
allowed-tools: Read, Write, Glob, Grep, AskUserQuestion
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
4. **Detect Incompleteness** - Scan for obvious ambiguities and ask for clarification (see below)
5. **Extract @agent Directives** - Use Grep tool to statically extract all directives (see below)
6. **Process @agent ask Directives** - Use AskUserQuestion immediately for any `@agent ask` directives
7. **Extract Project Parts** - Scan for service/system references, compare with registry
8. **Identify Artifacts** - Check if any analysis artifacts already exist that could be reused
9. **Process @agent require Directives** - Validate required artifacts exist or will be produced
10. **Identify Remaining Ambiguities** - Flag items needing clarification (excluding items with @agent answers)
11. **Ask Questions** - Use AskUserQuestion for critical decisions
12. **Apply Remaining @agent Directives** - Process notes, defer, priority, split, etc.
13. **Identify Modified Files** - For each implementation step, predict which files will be modified (see below)
14. **Output Brief** - Generate structured planning brief with all directives applied

### Step 4: Detect Incompleteness (Required)

Before proceeding with planning, scan the input for obvious ambiguities that would prevent creating actionable milestones. If found, use AskUserQuestion to clarify before continuing.

**Ambiguity Patterns to Detect:**

| Pattern | Example | Action |
|---------|---------|--------|
| Incomplete sentences | `- the button` | Ask: "What should 'the button' do?" |
| Missing context | `- fix it` | Ask: "What needs to be fixed?" |
| Vague references | `- update the thing` | Ask: "Which component should be updated?" |
| Unclear scope | `- improvements` | Ask: "What specific improvements?" |
| Trailing fragments | `- add feature for` | Ask: "Add feature for what purpose?" |
| Dangling items | `- see above` | Ask: "What does 'see above' refer to?" |
| Empty placeholders | `- TBD`, `- ???`, `- TODO` | Ask: "What should replace this placeholder?" |
| Orphaned sub-items | Sub-item without parent context | Ask: "What feature do these details belong to?" |

**Detection Heuristics:**

1. **Length check**: Items under 3 words are likely incomplete
2. **Verb check**: Items without action verbs may lack clear intent
3. **Reference check**: Pronouns (it, this, that) without clear antecedent
4. **Trailing punctuation**: Items ending in prepositions, articles, or conjunctions
5. **Placeholder markers**: TBD, TODO, ???, ..., [blank], etc.

**Report and Ask:**
```
Found 3 potentially incomplete items:

1. Line 15: "- the button"
   → Missing action or context

2. Line 23: "- fix it when"
   → Incomplete sentence

3. Line 45: "- TBD"
   → Placeholder needs definition

[AskUserQuestion: "I found some incomplete items. Please clarify:
1. What should 'the button' do?
2. What condition completes 'fix it when'?
3. What should replace 'TBD'?"]
```

**When NOT to flag:**
- Items with `@agent ask` directive (user already knows it needs detail)
- Items marked `(done)` (will be filtered anyway)
- Items in `ideas` section (intentionally vague)
- Items with sufficient context from parent grouping

### Step 5: Extract @agent Directives (Required)

**IMPORTANT**: Before processing the input, you MUST use the Grep tool to extract all `@agent` directives:

```
Grep pattern="@agent" path="<input-file>" output_mode="content" -B 1
```

This returns each `@agent` directive with 1 line of context before it (the parent item).

**Parse the output** to build a directive list:
```
[
  { "item": "don't show starter prompts if we restore a commit", "directive": "require", "value": "prompt_git_version_docs" },
  { "item": "generator call to /optimize...", "directive": "ask", "value": "how the history is tracked" },
  { "item": "add announcement bar", "directive": "note", "value": "needs legal review" }
]
```

**Report findings** before proceeding:
```
Found 4 @agent directives:
- 2x @agent require (artifacts: prompt_git_version_docs, auth_flow)
- 1x @agent ask (topic: history tracking)
- 1x @agent note (content: needs legal review)

Processing @agent ask directives first...
```

## Input Format Detection

The skill recognizes common backlog patterns:

```
## section          → Maturity/priority level (ready, needs planning, ideas)
*part*              → Logical grouping (UI area, service, feature set)
- item              → Individual feature/task
  - nested item     → Sub-task, detail, or acceptance criteria
(done)              → Completion marker (skip or filter)
@agent <directive>  → Inline instruction for the planning agent
```

## @agent Directives

The `@agent` marker provides inline instructions that the planning agent must recognize and follow. These directives guide how specific items should be handled during planning.

### Directive Types

| Pattern | Action |
|---------|--------|
| `@agent require <artifact>` | Add artifact to step's `requires` list |
| `@agent ask me about <topic>` | Use AskUserQuestion to get details |
| `@agent ask me for detail on <topic>` | Use AskUserQuestion to get specifics |
| `@agent requires documentation of <feature>` | Add documentation step with `produces` |
| `@agent split into <description>` | Split item into multiple milestones/steps |
| `@agent depends on <item>` | Add explicit dependency |
| `@agent defer` | Move to ideas/backlog section |
| `@agent priority <level>` | Override priority for this item |
| `@agent note: <text>` | Add to milestone/step notes |

### Processing @agent Directives

1. **Scan input** for `@agent` patterns (case-insensitive)
2. **Associate** each directive with its parent item (same line or indented below)
3. **Execute** directives during planning:
   - **require**: Look up artifact in registry, add to step's `requires`
   - **ask**: Queue question for AskUserQuestion, block until answered
   - **documentation**: Create analysis step with `produces: <feature>_docs`
   - **split**: Present splitting options to user
   - **depends on**: Add to `dependsOn` list
   - **defer**: Move item to ideas section
   - **note**: Capture in milestone/step notes

### Examples

**Input with @agent directives:**
```
- don't show starter prompts if we restore a commit @agent require prompt_git_version_docs
- generator call to /optimize should only include recent history (max 50)
  @agent ask me for detail on how the history is currently tracked
- add announcement bar @agent note: needs legal review for beta disclaimer
```

**Processing:**
1. First item: Look for `prompt_git_version_docs` artifact, add to requires
2. Second item: Ask user "How is the generator history currently tracked?" before generating steps
3. Third item: Add "needs legal review for beta disclaimer" to milestone notes

### @agent require

When you encounter `@agent require <artifact_id>`:

1. Check if artifact exists in registry
2. If exists: Add to step's `requires` list
3. If not exists:
   - Check if another step in this planning session `produces` it
   - If not found anywhere: Ask user where this artifact comes from

```yaml
# Input
- implement OAuth integration @agent require auth_flow_analysis

# Output step
- id: impl_oauth
  instruction: "implement OAuth integration"
  requires:
    - auth_flow_analysis  # Added from @agent directive
  projectParts: [api]
```

### @agent ask

When you encounter `@agent ask me about <topic>` or `@agent ask me for detail on <topic>`:

1. **Pause** processing of that item
2. **Use AskUserQuestion** to get the information
3. **Incorporate** the answer into the instruction or notes
4. **Continue** processing

```
# Input
- optimize prompt checkbox @agent ask me for detail on the optimization rules

# Action: AskUserQuestion
"The item 'optimize prompt checkbox' has a directive asking for details.
What are the optimization rules that should be applied?"

# User answers: "correct >1 comma, double space, prepend score tokens for pony/illustrious"

# Output step
- id: impl_optimize_checkbox
  instruction: "implement optimize prompt checkbox that corrects >1 comma, double space, and prepends score tokens for pony/illustrious ecosystems"
```

### @agent requires documentation of

When you encounter `@agent requires documentation of <feature>`:

1. **Create analysis step** that produces documentation artifact
2. **Add requires** for that artifact to the implementation step

```
# Input
- implement feature X @agent requires documentation of the Y system

# Output steps
- id: document_y_system
  instruction: "document the Y system architecture and interfaces"
  produces: y_system_docs
  projectParts: [relevant_part]

- id: impl_feature_x
  instruction: "implement feature X"
  requires:
    - y_system_docs
  dependsOn: [document_y_system]
```

### Multiple Directives

Items can have multiple @agent directives:

```
- complex feature @agent require api_docs @agent ask me about edge cases @agent note: high priority
```

Process all directives in order.

### Step 13: Identify Modified Files (Required)

For each implementation step, predict which files will be modified. This enables safe parallel execution by detecting file conflicts between steps.

**Why This Matters:**
When multiple steps run in parallel, they may try to modify the same file, causing overwrites. By declaring `modifies` upfront, the executor can:
- Run non-conflicting steps in parallel
- Serialize conflicting steps automatically
- Warn about potential merge issues

**Identification Strategies:**

| Feature Type | Typical Files Modified |
|--------------|------------------------|
| UI component | `index.html`, component JS, CSS |
| API endpoint | `handlers.go`, routes, types |
| State change | `app.js`, store, state files |
| Config change | Config files, constants |
| Test addition | Test files in e2e/tests dir |

**Heuristics for File Prediction:**

1. **Keyword Mapping:**
   - "button", "modal", "form" → HTML, component JS
   - "endpoint", "handler", "API" → backend handlers, routes
   - "state", "store" → state management files
   - "style", "theme" → CSS, style files

2. **Project Part Inference:**
   - `projectParts: [client]` → files in client directory
   - `projectParts: [api]` → files in api directory
   - `projectParts: [test]` → files in test directory

3. **Instruction Analysis:**
   - "modify X component" → X component file
   - "add to Y" → Y file
   - "update Z handler" → Z handler file

4. **Pattern Recognition:**
   - Similar past features → similar file patterns
   - Existing artifacts may document file locations

**When Uncertain:**
- Use broader estimates (list all potentially modified files)
- Mark as `# estimated` in comments
- The executor will treat missing `modifies` conservatively

**Example Analysis:**

```
Feature: "add download button to fullscreen view"

Analysis:
- "fullscreen view" → likely in fullscreen-related component
- "button" → UI element, needs HTML/JS
- "download" → may need click handler

Predicted modifies:
- client/fullscreen-viewer.js (UI component)
- client/index.html (if HTML changes needed)
- client/styles.css (if button styling needed)
```

**Output Format:**

```yaml
suggestedSteps:
  - id: impl_download
    instruction: "implement download button in fullscreen view"
    projectParts: [client]
    modifies:
      - client/fullscreen-viewer.js
      - client/index.html
```

**Conflict Detection in Brief:**

After generating all steps, scan for conflicts:

```
Analyzing file conflicts...

⚠️ Conflict detected:
  - impl_ui_polish modifies: [client/app.js]
  - impl_scroll_button modifies: [client/app.js]
  → These steps cannot run in parallel

Recommendation: Add dependency or serialize execution
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
        # Analysis steps typically don't modify files
      - id: impl_download
        instruction: "implement download button in fullscreen view"
        requires: [fullscreen_analysis]  # Uses the artifact
        dependsOn: [analyze_fullscreen]
        projectParts: [client]
        modifies:
          - client/fullscreen-viewer.js
          - client/index.html
      - id: impl_zoom
        instruction: "implement pinch/scroll zoom with viewport controls"
        requires: [fullscreen_analysis]
        dependsOn: [analyze_fullscreen]
        projectParts: [client]
        modifies:
          - client/fullscreen-viewer.js  # Conflict with impl_download!
          - client/touch-handlers.js
      - id: test_fullscreen
        instruction: "write tests for fullscreen view enhancements"
        dependsOn: [impl_download, impl_zoom]
        projectParts: [test]
        modifies:
          - client/e2e/fullscreen.spec.ts
    dependsOn: []
    notes: "User chose to group all fullscreen items together"
    fileConflicts:
      - files: [client/fullscreen-viewer.js]
        steps: [impl_download, impl_zoom]
        resolution: "Sequential execution required"

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
- [ ] Input was scanned for incomplete/ambiguous items and clarifications obtained
- [ ] All `@agent` directives were identified and processed
- [ ] All `@agent ask` questions were answered before proceeding
- [ ] All `@agent require` artifacts exist in registry or are produced by earlier steps
- [ ] All `@agent requires documentation of` items have corresponding analysis steps
- [ ] All `@agent note:` content is captured in milestone/step notes
- [ ] All detected project parts have paths (or are marked as registry updates)
- [ ] Analysis steps use `produces` for reusable insights
- [ ] Implementation steps use `requires` where applicable
- [ ] **Implementation steps have `modifies` field populated**
- [ ] **File conflicts between parallel steps are identified**
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

## Example Session with @agent Directives

**User:** `/milestone-plan /path/to/scratch.txt`

**scratch.txt contents:**
```
## ready

*generator*
- don't show starter prompts if we restore a commit @agent require prompt_git_version_docs
- generator call to /optimize should only include recent history (max 50)
  @agent ask me for detail on how the history is tracked
- add announcement bar @agent note: needs legal review

*tools*
- prompt copilot @agent requires documentation of the dspy service
```

**Assistant:**

Step 1: Check registry
```
[Read .accelyst/registry.yaml]

Registry contains:
- Project parts: client, api, dspy-service, test
- Artifacts: api_patterns
```

Step 2-3: Read and detect format
```
[Read /path/to/scratch.txt]

Detected: 1 section (ready), 2 parts (generator, tools), 4 items
```

Step 4: Extract @agent directives using Grep
```
[Grep pattern="@agent" path="/path/to/scratch.txt" output_mode="content" -B 1]

Results:
  3: - don't show starter prompts if we restore a commit @agent require prompt_git_version_docs
  --
  4: - generator call to /optimize should only include recent history (max 50)
  5:   @agent ask me for detail on how the history is tracked
  --
  6: - add announcement bar @agent note: needs legal review
  --
  9: - prompt copilot @agent requires documentation of the dspy service
```

Parsed directives:
| Item | Directive | Value |
|------|-----------|-------|
| don't show starter prompts... | require | prompt_git_version_docs |
| generator call to /optimize... | ask | how the history is tracked |
| add announcement bar | note | needs legal review |
| prompt copilot | requires documentation of | the dspy service |

Step 5: Process @agent ask directives first

[AskUserQuestion: "How is the generator history currently tracked? (from @agent directive on 'generator call to /optimize')"]

**User answers:** "History is stored in localStorage as an array of prompt objects with timestamps"

Step 8: Process @agent require directives

Checking artifact `prompt_git_version_docs`... NOT FOUND in registry.

[AskUserQuestion: "The item 'don't show starter prompts' requires artifact 'prompt_git_version_docs' which doesn't exist. Options: 1) Create a documentation step that produces it (Recommended), 2) Skip this requirement, 3) You'll provide the artifact"]

**User answers:** "Create a documentation step"

Step 11: Apply remaining directives

- `@agent note: needs legal review` → Added to announcement_bar milestone notes
- `@agent requires documentation of the dspy service` → Created document_dspy_service step

```
All @agent directives processed. Creating planning brief...

Brief includes:
- 1 new analysis step: document_prompt_git_version (produces: prompt_git_version_docs)
- 1 new analysis step: document_dspy_service (produces: dspy_service_docs)
- impl_starter_prompts requires: [prompt_git_version_docs]
- impl_prompt_copilot requires: [dspy_service_docs]
- impl_optimize_history instruction updated with localStorage details
- announcement_bar milestone has note: "needs legal review"

Brief saved to: planning-brief.yaml
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

## Related Skills

- `/milestone-orchestrate` - **Full-cycle**: runs plan → convert → execute automatically
- `/milestone-convert` - Convert planning brief → epic YAML (next step)
- `/milestone-execute` - Execute epic via subagents
- `/accelyst-guide` - CLI operations and troubleshooting
