# Subagent Execution Templates

This document defines the prompt templates used when spawning subagents for step execution.

## Template Variables

| Variable | Source | Description |
|----------|--------|-------------|
| `{step.id}` | Epic YAML | Step identifier |
| `{step.instruction}` | Epic YAML | The action to perform |
| `{step.produces}` | Epic YAML | Artifact ID to create (optional) |
| `{step.requires}` | Epic YAML | List of artifact IDs needed |
| `{step.dependsOn}` | Epic YAML | List of prerequisite step IDs |
| `{step.projectParts}` | Epic YAML | List of project part names |
| `{part.name}` | Registry | Project part identifier |
| `{part.directoryAbs}` | Registry | Absolute path to component |
| `{part.documentationAbs}` | Registry | Path to docs (optional) |
| `{artifact.content}` | Artifact file | Full markdown content |
| `{epic.name}` | Epic YAML | Epic display name |
| `{milestone.name}` | Epic YAML | Milestone display name |

## Base Template Structure

All subagent prompts follow this structure:

```markdown
# Task: {milestone.name} / {step.id}

## Project Scope

{for each part in step.projectParts}
### {part.name}
- Directory: {part.directoryAbs}
{if part.documentationAbs}- Documentation: {part.documentationAbs}{/if}
{/for}

{if step.requires}
## Required Context

{for each artifact_id in step.requires}
### Artifact: {artifact_id}

{artifact.content}

---
{/for}
{/if}

{if step.dependsOn}
## Prior Steps

This step depends on completion of: {step.dependsOn joined by ", "}

Review their outputs if relevant to your task.
{/if}

## Instruction

{step.instruction}

{if step.produces}
## Output Required

Save your analysis to: `.accelyst/artifacts/{step.produces}.md`

{artifact_format_guidance}
{/if}

## Guidelines

{task_specific_guidelines}
```

## Analysis Step Template

Used when step has `produces` field (creating an artifact).

**Subagent type:** `Explore`

```markdown
# Analysis Task: {step.id}

You are analyzing code to produce insights for subsequent implementation steps.

## Project Scope

{for each part in step.projectParts}
### {part.name}
- Directory: {part.directoryAbs}
{if part.documentationAbs}- Documentation: {part.documentationAbs}{/if}

Explore this directory to understand the component's structure, patterns, and interfaces.
{/for}

{if step.requires}
## Prior Analysis

Review this existing analysis for context:

{for each artifact_id in step.requires}
### {artifact_id}

{artifact.content}

---
{/for}
{/if}

## Analysis Task

{step.instruction}

## Output Required

Save your analysis to: `.accelyst/artifacts/{step.produces}.md`

Structure your analysis as follows:

```markdown
# {step.produces}

## Overview
Brief summary of findings (2-3 sentences)

## Key Files
| File | Purpose |
|------|---------|
| path/to/file.js | Description of what it does |

## Patterns Identified
- Pattern 1: Description
- Pattern 2: Description

## Key Interfaces
```typescript
// Important types, interfaces, or function signatures
```

## State Management
How state flows through the component (if applicable)

## Extension Points
Where and how new functionality should be added

## Recommendations
Specific suggestions for implementation
```

## Guidelines

- Focus on information useful for implementation
- Include specific file paths and line numbers
- Document patterns that should be followed
- Note any gotchas or edge cases
- Keep the artifact concise but comprehensive
- Use code snippets to illustrate patterns
```

## Implementation Step Template

Used when step modifies code (no `produces`, has implementation instruction).

**Subagent type:** Based on projectParts (see SKILL.md)

```markdown
# Implementation Task: {step.id}

You are implementing a feature based on prior analysis and requirements.

## Project Scope

{for each part in step.projectParts}
### {part.name}
- Directory: {part.directoryAbs}
{if part.documentationAbs}- Documentation: {part.documentationAbs}{/if}
{/for}

{if step.requires}
## Context from Analysis

Use this analysis to guide your implementation:

{for each artifact_id in step.requires}
### {artifact_id}

{artifact.content}

---
{/for}
{/if}

{if step.dependsOn}
## Prior Step Results

This step builds on: {step.dependsOn joined by ", "}

Their work is complete. Build on what they created.
{/if}

## Implementation Task

{step.instruction}

## Guidelines

- Follow patterns identified in the analysis artifacts
- Match existing code style and conventions
- Create focused, minimal changes
- Add appropriate error handling
- Include inline comments only where logic isn't self-evident
- Test your changes manually if possible
- Report what you implemented and any issues encountered

## Completion

When done, summarize:
1. Files modified/created
2. Key changes made
3. Any assumptions or decisions
4. Known limitations or TODOs
```

## Test Step Template

Used when `test` appears in projectParts.

**Subagent type:** `quality-engineer`

```markdown
# Testing Task: {step.id}

You are writing tests to verify functionality.

## Test Scope

{for each part in step.projectParts}
### {part.name}
- Directory: {part.directoryAbs}
{if part.documentationAbs}- Documentation: {part.documentationAbs}{/if}
{/for}

{if step.requires}
## Implementation Context

{for each artifact_id in step.requires}
### {artifact_id}

{artifact.content}

---
{/for}
{/if}

{if step.dependsOn}
## Implementation to Test

This tests the work from: {step.dependsOn joined by ", "}

Review what was implemented to understand what needs testing.
{/if}

## Testing Task

{step.instruction}

## Guidelines

- Write tests that verify the core functionality
- Include edge cases and error scenarios
- Follow existing test patterns in the project
- Ensure tests are deterministic and isolated
- Use meaningful test names that describe behavior
- Run tests to verify they pass

## Test Categories to Consider

1. **Happy Path** - Normal expected usage
2. **Edge Cases** - Boundary conditions, empty inputs
3. **Error Handling** - Invalid inputs, failure scenarios
4. **Integration** - Component interactions (if applicable)

## Completion

When done, report:
1. Test files created/modified
2. Number of test cases
3. Test execution results
4. Any issues found during testing
```

## Infrastructure Step Template

Used when `infra` appears in projectParts.

**Subagent type:** `devops-architect`

```markdown
# Infrastructure Task: {step.id}

You are modifying infrastructure or deployment configuration.

## Infrastructure Scope

{for each part in step.projectParts}
### {part.name}
- Directory: {part.directoryAbs}
{if part.documentationAbs}- Documentation: {part.documentationAbs}{/if}
{/for}

{if step.requires}
## Context

{for each artifact_id in step.requires}
### {artifact_id}

{artifact.content}

---
{/for}
{/if}

## Infrastructure Task

{step.instruction}

## Guidelines

- Validate configuration syntax before applying
- Consider rollback procedures
- Document any manual steps required
- Test in non-production first if possible
- Follow GitOps practices if applicable

## Safety Checks

Before making changes:
- [ ] Backup current configuration
- [ ] Understand blast radius of changes
- [ ] Have rollback plan ready

## Completion

When done, report:
1. Configuration files modified
2. Resources affected
3. Deployment steps (if any)
4. Verification steps
```

## Security Step Template

Used when instruction contains security-related keywords.

**Subagent type:** `security-engineer`

```markdown
# Security Task: {step.id}

You are implementing or reviewing security-related functionality.

## Scope

{for each part in step.projectParts}
### {part.name}
- Directory: {part.directoryAbs}
{if part.documentationAbs}- Documentation: {part.documentationAbs}{/if}
{/for}

{if step.requires}
## Security Context

{for each artifact_id in step.requires}
### {artifact_id}

{artifact.content}

---
{/for}
{/if}

## Security Task

{step.instruction}

## Guidelines

- Follow OWASP best practices
- Validate all inputs
- Use parameterized queries for database operations
- Implement proper authentication/authorization
- Handle sensitive data appropriately
- Log security-relevant events
- Consider rate limiting where appropriate

## Security Checklist

- [ ] No hardcoded secrets or credentials
- [ ] Input validation implemented
- [ ] Output encoding for XSS prevention
- [ ] Authentication properly enforced
- [ ] Authorization checks in place
- [ ] Sensitive data protected

## Completion

When done, report:
1. Security measures implemented
2. Potential risks mitigated
3. Remaining security considerations
4. Recommended follow-up actions
```

## Subagent Type Selection Logic

```python
def select_subagent_type(step):
    # Analysis steps use Explore agent
    if step.produces:
        return "Explore"

    # Check projectParts for specialization
    parts = step.projectParts or []

    if "test" in parts:
        return "quality-engineer"

    if "infra" in parts:
        return "devops-architect"

    if "api" in parts or "backend" in parts:
        return "backend-architect"

    if "client" in parts or "frontend" in parts:
        return "frontend-architect"

    # Check instruction for security keywords
    instruction = step.instruction.lower()
    security_keywords = ["security", "auth", "permission", "credential", "encrypt"]
    if any(kw in instruction for kw in security_keywords):
        return "security-engineer"

    # Default to general-purpose
    return "general-purpose"
```

## Prompt Assembly Process

```python
def build_step_prompt(step, registry, artifacts):
    # 1. Select template based on step type
    if step.produces:
        template = ANALYSIS_TEMPLATE
    elif "test" in step.projectParts:
        template = TEST_TEMPLATE
    elif "infra" in step.projectParts:
        template = INFRA_TEMPLATE
    else:
        template = IMPLEMENTATION_TEMPLATE

    # 2. Resolve project parts from registry
    parts = []
    for part_name in step.projectParts:
        part = registry.projectParts.find(name=part_name)
        parts.append(part)

    # 3. Load required artifacts
    artifact_contents = {}
    for artifact_id in step.requires:
        content = read_file(f".accelyst/artifacts/{artifact_id}.md")
        artifact_contents[artifact_id] = content

    # 4. Render template with variables
    prompt = template.render(
        step=step,
        parts=parts,
        artifacts=artifact_contents,
        milestone=milestone,
        epic=epic
    )

    return prompt
```

## Example Rendered Prompt

For step `analyze_fullscreen` in `ready-features/fullscreen_download`:

```markdown
# Analysis Task: analyze_fullscreen

You are analyzing code to produce insights for subsequent implementation steps.

## Project Scope

### client
- Directory: /home/zach/workspace/promptver/client

Explore this directory to understand the component's structure, patterns, and interfaces.

## Analysis Task

analyze the fullscreen view component implementation

## Output Required

Save your analysis to: `.accelyst/artifacts/fullscreen_view_analysis.md`

Structure your analysis as follows:

# fullscreen_view_analysis

## Overview
Brief summary of findings (2-3 sentences)

## Key Files
| File | Purpose |
|------|---------|
| path/to/file.js | Description of what it does |

## Patterns Identified
- Pattern 1: Description
- Pattern 2: Description

## Key Interfaces
```typescript
// Important types, interfaces, or function signatures
```

## State Management
How state flows through the component (if applicable)

## Extension Points
Where and how new functionality should be added

## Recommendations
Specific suggestions for implementation

## Guidelines

- Focus on information useful for implementation
- Include specific file paths and line numbers
- Document patterns that should be followed
- Note any gotchas or edge cases
- Keep the artifact concise but comprehensive
- Use code snippets to illustrate patterns
```
