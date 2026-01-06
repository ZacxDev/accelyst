package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

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
	ID               int      `yaml:"id"`
	Instruction      string   `yaml:"instruction"`
	DependsOnStepIDs []int    `yaml:"dependsOnStepIds,omitempty"`
	ProjectParts     []string `yaml:"projectParts,omitempty"`
}

// Milestone represents a collection of steps working toward a goal
type Milestone struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

// Config represents the full project configuration
type Config struct {
	ProjectParts []ProjectPart `yaml:"projectParts"`
	Milestones   []Milestone   `yaml:"milestones"`
}

// DAG performs topological sort on steps using Kahn's algorithm
func topologicalSort(steps []Step) ([]Step, error) {
	// Build adjacency list and in-degree map
	stepByID := make(map[int]Step)
	inDegree := make(map[int]int)
	adjacency := make(map[int][]int) // step -> steps that depend on it

	for _, step := range steps {
		stepByID[step.ID] = step
		inDegree[step.ID] = len(step.DependsOnStepIDs)

		for _, depID := range step.DependsOnStepIDs {
			adjacency[depID] = append(adjacency[depID], step.ID)
		}
	}

	// Find all steps with no dependencies (in-degree 0)
	var queue []int
	for _, step := range steps {
		if inDegree[step.ID] == 0 {
			queue = append(queue, step.ID)
		}
	}

	// Sort initial queue for deterministic output
	sort.Ints(queue)

	var sorted []Step
	for len(queue) > 0 {
		// Take first element
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, stepByID[current])

		// Get all steps that depend on current, sort for deterministic output
		dependents := adjacency[current]
		sort.Ints(dependents)

		for _, depID := range dependents {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
				sort.Ints(queue)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(steps) {
		return nil, fmt.Errorf("cycle detected in step dependencies")
	}

	return sorted, nil
}

// generatePromptFragment creates the prompt fragment for a step
func generatePromptFragment(step Step, projectParts map[string]ProjectPart) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("%d:", step.ID))

	// Steps with no dependencies can run in parallel via subagents
	if len(step.DependsOnStepIDs) == 0 {
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
	if len(step.DependsOnStepIDs) > 0 {
		var depStrs []string
		for _, depID := range step.DependsOnStepIDs {
			depStrs = append(depStrs, fmt.Sprintf("%d", depID))
		}
		parts = append(parts, fmt.Sprintf("read results from steps %s then", strings.Join(depStrs, ", ")))
	}

	// Add instruction
	parts = append(parts, step.Instruction)

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

	// Process each milestone
	for i, milestone := range config.Milestones {
		if i > 0 {
			fmt.Println("\n---\n")
		}

		fmt.Printf("# Milestone: %s\n\n", milestone.Name)

		prompt, err := processMilestone(milestone, projectParts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error processing milestone: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(prompt)
	}
}
