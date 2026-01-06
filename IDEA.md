Accelyst aims to create a structured representation of a project and it's evolution to make it easy for AI Agents to autonomously go from highest-level feature description to production deployed implementation with zero human intervention.

It does this by parsing the following:

```
ProjectPart {
  name: string
  directoryAbs: string
  documentationAbs: string
}

Milestone {
 name: string
  steps: {
    id: number
    instruction: string
    dependsOnStepIds: number[]
    projectParts: string[]
  }[]
}

exampleMilestone:
  name: "Update UI Style"
  steps:
    - id: 1
      instruction: "use context7 and search to pull the tailwind v4 documentation"
      dependsOnStepIds: []
    - id: 2
      instruction: "use context7 and search to pull the lit webcomponents documentation"
      dependsOnStepIds: []
    - id: 3
      instruction: "analyze the client implementation"
      dependsOnStepIds: []
    - id: 4
      instruction: "migrate the existing client html/css to use tailwind and lit webcomponents, update the styling and layout to be simple and delightful"
      dependsOnStepIds:
        - 1
        - 2
        - 3
```

by:

- parse projectParts and milestones
- in order, for each milestone:
  - create a DAG to sort the steps
  - in DAG order, for each step:
      - resolve projectParts and dependsOnStepIds
      - generate a 'StepExecutionPromptFragment' `${id}: ${projectParts ? `Analyze ${each p projectParts p.name (p.directoryAbs) p.documentationAbs ? `(docs: p.documentationAbs)` : ''} then ` : ''}${dependsOnStepIds ? `read results from steps ${each s dependsOnStepIds s.id} then ` : ''}${instruction}`
  - concat the StepExecutionPromptFragments and print

