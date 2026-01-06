# Accelyst

A structured project evolution tool that generates execution prompts for AI agents to autonomously implement features from high-level descriptions.

## Overview

Accelyst parses project structure and milestone definitions, then generates topologically-sorted execution prompts. Steps without dependencies are marked for parallel subagent execution.

## Installation

```bash
go build -o accelyst
```

## Usage

```bash
./accelyst                      # uses accelyst.yaml
./accelyst -config custom.yaml  # custom config file
```

## Configuration Format

```yaml
projectParts:
  - name: client
    directoryAbs: /path/to/client
    documentationAbs: /path/to/docs/client.md  # optional
  - name: server
    directoryAbs: /path/to/server

milestones:
  - name: "Feature Name"
    steps:
      - id: 1
        instruction: "research the documentation"
        dependsOnStepIds: []
        projectParts: []
      - id: 2
        instruction: "analyze the existing implementation"
        dependsOnStepIds: []
        projectParts:
          - client
      - id: 3
        instruction: "implement the feature"
        dependsOnStepIds:
          - 1
          - 2
        projectParts:
          - client
          - server
```

## Output Format

```
# Milestone: Feature Name

1: use a subagent to research the documentation
2: use a subagent to Analyze client (/path/to/client) then analyze the existing implementation
3: Analyze client (/path/to/client), server (/path/to/server) then read results from steps 1, 2 then implement the feature
```

- Steps with no dependencies get `use a subagent to` prefix (parallelizable)
- Steps with dependencies include `read results from steps X, Y`
- Steps with project parts include `Analyze <name> (<path>) (docs: <path>)`

## How It Works

1. Parse `projectParts` and `milestones` from YAML
2. For each milestone:
   - Build a DAG from step dependencies
   - Topologically sort steps (Kahn's algorithm)
   - Generate prompt fragments with resolved references
3. Output concatenated prompts per milestone
