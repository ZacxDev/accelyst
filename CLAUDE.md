# Accelyst

AI-native project iteration interface. Transforms feature backlogs into execution-ready prompts.

## Documentation Index

| File | Read When |
|------|-----------|
| `README.md` | Overview, usage, configuration format |
| `SPEC.md` | Schema details, constraints, JSON schema, execution semantics |
| `main.go` | Implementation reference, modifying behavior |
| `main_test.go` | Test patterns, validation edge cases |

## Key Concepts

- **ProjectPart**: Codebase component with path and optional docs
- **Milestone**: Feature goal with `id`, `dependsOn`, and steps
- **Step**: Atomic action with `instruction`, `acceptanceCriteria`, `tier`

## Tiers

- `ai` (default): Autonomous execution
- `human`: Requires human action

## Output Format

```
step_id: [HUMAN]? [subagent prefix]? [project parts]? [dependencies]? instruction [done when: criteria]?
```
