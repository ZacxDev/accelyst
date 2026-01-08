package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// ProjectPart represents a component of the project with its location and documentation
type ProjectPart struct {
	Name             string `yaml:"name"`
	DirectoryAbs     string `yaml:"directoryAbs"`
	DocumentationAbs string `yaml:"documentationAbs,omitempty"`
}

// Step represents a single action within a milestone
type Step struct {
	ID                 string   `yaml:"id"`
	Instruction        string   `yaml:"instruction"`
	DependsOn          []string `yaml:"dependsOn,omitempty"`
	ProjectParts       []string `yaml:"projectParts,omitempty"`
	AcceptanceCriteria []string `yaml:"acceptanceCriteria,omitempty"`
	Tier               string   `yaml:"tier,omitempty"` // "ai" (default) or "human"
}

// Milestone represents a collection of steps working toward a goal
type Milestone struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
	Steps     []Step   `yaml:"steps"`
}

// Config represents the full project configuration
type Config struct {
	ProjectParts []ProjectPart `yaml:"projectParts"`
	Milestones   []Milestone   `yaml:"milestones"`
}

// JSON Schema for validation (from SPEC.md)
const configSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["projectParts", "milestones"],
  "properties": {
    "projectParts": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "directoryAbs"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "directoryAbs": { "type": "string", "minLength": 1 },
          "documentationAbs": { "type": "string" }
        }
      }
    },
    "milestones": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "name", "steps"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "name": { "type": "string", "minLength": 1 },
          "dependsOn": {
            "type": "array",
            "items": { "type": "string", "minLength": 1 }
          },
          "steps": {
            "type": "array",
            "minItems": 1,
            "items": {
              "type": "object",
              "required": ["id", "instruction"],
              "properties": {
                "id": { "type": "string", "minLength": 1 },
                "instruction": { "type": "string", "minLength": 1 },
                "dependsOn": {
                  "type": "array",
                  "items": { "type": "string", "minLength": 1 }
                },
                "projectParts": {
                  "type": "array",
                  "items": { "type": "string" }
                },
                "acceptanceCriteria": {
                  "type": "array",
                  "items": { "type": "string", "minLength": 1 }
                },
                "tier": {
                  "type": "string",
                  "enum": ["ai", "human"]
                }
              }
            }
          }
        }
      }
    }
  }
}`

// validateConfig validates the config against the JSON schema
func validateConfig(data []byte) error {
	// Parse YAML to generic interface for JSON conversion
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(convertYAMLToJSON(raw))
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Validate against schema
	schemaLoader := gojsonschema.NewStringLoader(configSchema)
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errs []string
		for _, err := range result.Errors() {
			errs = append(errs, fmt.Sprintf("  - %s", err.String()))
		}
		return fmt.Errorf("config validation failed:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

// convertYAMLToJSON converts YAML-parsed data to JSON-compatible format
// (handles map[interface{}]interface{} -> map[string]interface{})
func convertYAMLToJSON(v interface{}) interface{} {
	switch v := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			m[fmt.Sprintf("%v", k)] = convertYAMLToJSON(val)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = convertYAMLToJSON(val)
		}
		return v
	default:
		return v
	}
}

// DAG performs topological sort on steps using Kahn's algorithm
func topologicalSort(steps []Step) ([]Step, error) {
	// Build adjacency list and in-degree map
	stepByID := make(map[string]Step)
	inDegree := make(map[string]int)
	adjacency := make(map[string][]string) // step -> steps that depend on it

	// Initialize maps
	for _, step := range steps {
		stepByID[step.ID] = step
		inDegree[step.ID] = 0 // Ensure all steps are in the map
	}

	for _, step := range steps {
		inDegree[step.ID] = len(step.DependsOn)

		for _, depID := range step.DependsOn {
			if _, exists := stepByID[depID]; !exists {
				return nil, fmt.Errorf("step %q depends on non-existent step %q", step.ID, depID)
			}
			adjacency[depID] = append(adjacency[depID], step.ID)
		}
	}

	// Find all steps with no dependencies (in-degree 0)
	var queue []string
	for _, step := range steps {
		if inDegree[step.ID] == 0 {
			queue = append(queue, step.ID)
		}
	}

	// Sort initial queue for deterministic output
	sort.Strings(queue)

	var sorted []Step
	for len(queue) > 0 {
		// Take first element
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, stepByID[current])

		// Get all steps that depend on current, sort for deterministic output
		dependents := adjacency[current]
		sort.Strings(dependents)

		for _, depID := range dependents {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
				sort.Strings(queue)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(steps) {
		return nil, fmt.Errorf("cycle detected in step dependencies")
	}

	return sorted, nil
}

// topologicalSortMilestones performs topological sort on milestones using Kahn's algorithm
func topologicalSortMilestones(milestones []Milestone) ([]Milestone, error) {
	// Build adjacency list and in-degree map
	milestoneByID := make(map[string]Milestone)
	inDegree := make(map[string]int)
	adjacency := make(map[string][]string) // milestone -> milestones that depend on it

	// Initialize maps
	for _, m := range milestones {
		if m.ID == "" {
			return nil, fmt.Errorf("milestone %q is missing required 'id' field", m.Name)
		}
		if _, exists := milestoneByID[m.ID]; exists {
			return nil, fmt.Errorf("duplicate milestone id %q", m.ID)
		}
		milestoneByID[m.ID] = m
		inDegree[m.ID] = 0
	}

	for _, m := range milestones {
		inDegree[m.ID] = len(m.DependsOn)

		for _, depID := range m.DependsOn {
			if _, exists := milestoneByID[depID]; !exists {
				return nil, fmt.Errorf("milestone %q depends on non-existent milestone %q", m.ID, depID)
			}
			adjacency[depID] = append(adjacency[depID], m.ID)
		}
	}

	// Find all milestones with no dependencies (in-degree 0)
	var queue []string
	for _, m := range milestones {
		if inDegree[m.ID] == 0 {
			queue = append(queue, m.ID)
		}
	}

	// Sort initial queue for deterministic output
	sort.Strings(queue)

	var sorted []Milestone
	for len(queue) > 0 {
		// Take first element
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, milestoneByID[current])

		// Get all milestones that depend on current, sort for deterministic output
		dependents := adjacency[current]
		sort.Strings(dependents)

		for _, depID := range dependents {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
				sort.Strings(queue)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(milestones) {
		return nil, fmt.Errorf("cycle detected in milestone dependencies")
	}

	return sorted, nil
}

// generatePromptFragment creates the prompt fragment for a step
func generatePromptFragment(step Step, projectParts map[string]ProjectPart) string {
	var parts []string

	// Add tier prefix for human steps
	if step.Tier == "human" {
		parts = append(parts, fmt.Sprintf("%s: [HUMAN]", step.ID))
	} else {
		parts = append(parts, fmt.Sprintf("%s:", step.ID))
	}

	// Steps with no dependencies can run in parallel via subagents (only for AI tier)
	if len(step.DependsOn) == 0 && step.Tier != "human" {
		parts = append(parts, "use a subagent to")
	}

	// Add project parts analysis section
	if len(step.ProjectParts) > 0 {
		var partDescriptions []string
		for _, partName := range step.ProjectParts {
			if part, ok := projectParts[partName]; ok {
				desc := fmt.Sprintf("%s (%s)", part.Name, part.DirectoryAbs)
				if part.DocumentationAbs != "" {
					desc += fmt.Sprintf(" (docs: %s)", part.DocumentationAbs)
				}
				partDescriptions = append(partDescriptions, desc)
			}
		}
		if len(partDescriptions) > 0 {
			parts = append(parts, fmt.Sprintf("Analyze %s then", strings.Join(partDescriptions, ", ")))
		}
	}

	// Add dependencies section
	if len(step.DependsOn) > 0 {
		parts = append(parts, fmt.Sprintf("read results from steps %s then", strings.Join(step.DependsOn, ", ")))
	}

	// Add instruction
	parts = append(parts, step.Instruction)

	// Add acceptance criteria if present
	if len(step.AcceptanceCriteria) > 0 {
		criteria := strings.Join(step.AcceptanceCriteria, "; ")
		parts = append(parts, fmt.Sprintf("[done when: %s]", criteria))
	}

	return strings.Join(parts, " ")
}

// processMilestone processes a single milestone and returns the combined prompt
func processMilestone(milestone Milestone, projectParts map[string]ProjectPart) (string, error) {
	// Topologically sort steps
	sortedSteps, err := topologicalSort(milestone.Steps)
	if err != nil {
		return "", fmt.Errorf("milestone %q: %w", milestone.Name, err)
	}

	var fragments []string
	for _, step := range sortedSteps {
		fragment := generatePromptFragment(step, projectParts)
		fragments = append(fragments, fragment)
	}

	return strings.Join(fragments, "\n"), nil
}

func main() {
	configPath := flag.String("config", "accelyst.yaml", "path to configuration file")
	flag.Parse()

	// Read config file
	data, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
		os.Exit(1)
	}

	// Validate against JSON schema
	if err := validateConfig(data); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config: %v\n", err)
		os.Exit(1)
	}

	// Build project parts lookup
	projectParts := make(map[string]ProjectPart)
	for _, part := range config.ProjectParts {
		projectParts[part.Name] = part
	}

	// Topologically sort milestones
	sortedMilestones, err := topologicalSortMilestones(config.Milestones)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sorting milestones: %v\n", err)
		os.Exit(1)
	}

	// Process each milestone in sorted order
	for i, milestone := range sortedMilestones {
		if i > 0 {
			fmt.Print("\n---\n\n")
		}

		// Build milestone header with dependencies
		header := fmt.Sprintf("# Milestone: %s", milestone.Name)
		if len(milestone.DependsOn) > 0 {
			header += fmt.Sprintf(" (depends on: %s)", strings.Join(milestone.DependsOn, ", "))
		}
		fmt.Println(header)
		fmt.Println()

		prompt, err := processMilestone(milestone, projectParts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error processing milestone: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(prompt)
	}
}
