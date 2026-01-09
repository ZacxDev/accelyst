# Accelyst Project Overview

## Purpose
Accelyst is an AI-native project iteration interface that transforms feature backlogs into execution-ready prompts. It generates prompts with full codebase context including directories, documentation paths, dependency ordering, and artifact context for AI agent consumption.

## Key Features
- **Rich context**: Prompts reference exact directories and documentation
- **Dependency-aware**: Milestones and steps are topologically sorted via DAG (Kahn's algorithm)
- **Parallel-ready**: Independent work marked for concurrent subagent execution
- **Knowledge persistence**: Analysis artifacts accumulate across iterations
- **Multiple workstreams**: Support for concurrent epics

## Tech Stack
- **Language**: Go 1.25.5
- **Dependencies**:
  - `gopkg.in/yaml.v3` - YAML parsing
  - `github.com/xeipuuv/gojsonschema` - JSON schema validation

## Architecture (Split Configuration Model)
```
project/
├── .accelyst/
│   ├── registry.yaml      # Persistent: project parts + artifacts
│   ├── epics/             # Ephemeral: milestone collections
│   │   ├── feature-a.yaml
│   │   └── feature-b.yaml
│   └── artifacts/         # Knowledge outputs
│       └── analysis.md
```

- **Registry** (`.accelyst/registry.yaml`): Persistent project knowledge
- **Epics** (`.accelyst/epics/*.yaml`): Disposable milestone collections
- **Artifacts** (`.accelyst/artifacts/*.md`): Analysis outputs that persist across iterations

## Core Concepts
- **ProjectPart**: Codebase component with path and optional docs
- **Milestone**: Feature goal with `id`, `dependsOn`, and steps
- **Step**: Atomic action with `instruction`, `projectParts`, `produces`, `requires`
- **Artifact**: Markdown file capturing analysis outputs that persist across iterations
