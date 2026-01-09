package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// Registry types

// ProjectPart represents a component of the project with its location and documentation
type ProjectPart struct {
	Name             string `yaml:"name"`
	DirectoryAbs     string `yaml:"directoryAbs"`
	DocumentationAbs string `yaml:"documentationAbs,omitempty"`
	Description      string `yaml:"description,omitempty"`
}

// Artifact represents a registered analysis artifact
type Artifact struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	ProducedBy  string `yaml:"producedBy"`
	CreatedAt   string `yaml:"createdAt"`
}

// Registry represents the persistent project knowledge
type Registry struct {
	Version      int           `yaml:"version"`
	ProjectParts []ProjectPart `yaml:"projectParts"`
	Artifacts    []Artifact    `yaml:"artifacts,omitempty"`
}

// Epic types

// Step represents a single action within a milestone
type Step struct {
	ID           string   `yaml:"id"`
	Instruction  string   `yaml:"instruction"`
	DependsOn    []string `yaml:"dependsOn,omitempty"`
	ProjectParts []string `yaml:"projectParts,omitempty"`
	Produces     string   `yaml:"produces,omitempty"`
	Requires     []string `yaml:"requires,omitempty"`
	Modifies     []string `yaml:"modifies,omitempty"`
}

// Milestone represents a collection of steps working toward a goal
type Milestone struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
	Steps     []Step   `yaml:"steps"`
}

// Epic represents a collection of related milestones
type Epic struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description,omitempty"`
	Priority    string      `yaml:"priority,omitempty"`
	Milestones  []Milestone `yaml:"milestones"`
	filename    string      // internal: source filename (without extension)
}

// Legacy Config type for migration
type LegacyConfig struct {
	ProjectParts []ProjectPart `yaml:"projectParts"`
	Milestones   []Milestone   `yaml:"milestones"`
}

// JSON Schemas

const registrySchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "projectParts"],
  "properties": {
    "version": { "type": "integer" },
    "projectParts": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "directoryAbs"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "directoryAbs": { "type": "string", "minLength": 1 },
          "documentationAbs": { "type": "string" },
          "description": { "type": "string" }
        }
      }
    },
    "artifacts": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "description", "producedBy", "createdAt"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "description": { "type": "string", "minLength": 1 },
          "producedBy": { "type": "string", "minLength": 1 },
          "createdAt": { "type": "string" }
        }
      }
    }
  }
}`

const epicSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "milestones"],
  "properties": {
    "name": { "type": "string", "minLength": 1 },
    "description": { "type": "string" },
    "priority": { "type": "string", "enum": ["ready", "in_progress", "blocked", "ideas"] },
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
                "produces": { "type": "string" },
                "requires": {
                  "type": "array",
                  "items": { "type": "string" }
                },
                "modifies": {
                  "type": "array",
                  "items": { "type": "string" }
                }
              }
            }
          }
        }
      }
    }
  }
}`

// Validation functions

func validateYAML(data []byte, schema string) error {
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	jsonData, err := json.Marshal(convertYAMLToJSON(raw))
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
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
		return fmt.Errorf("validation failed:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

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

// Registry functions

func loadRegistry(basePath string) (*Registry, error) {
	registryPath := filepath.Join(basePath, ".accelyst", "registry.yaml")

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	if err := validateYAML(data, registrySchema); err != nil {
		return nil, fmt.Errorf("registry validation: %w", err)
	}

	var registry Registry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return &registry, nil
}

func saveRegistry(basePath string, registry *Registry) error {
	registryPath := filepath.Join(basePath, ".accelyst", "registry.yaml")

	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// Epic functions

func loadEpics(basePath string) ([]Epic, error) {
	epicsDir := filepath.Join(basePath, ".accelyst", "epics")

	entries, err := os.ReadDir(epicsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read epics directory: %w", err)
	}

	var epics []Epic
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		epicPath := filepath.Join(epicsDir, entry.Name())
		data, err := os.ReadFile(epicPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read epic %s: %w", entry.Name(), err)
		}

		if err := validateYAML(data, epicSchema); err != nil {
			return nil, fmt.Errorf("epic %s: %w", entry.Name(), err)
		}

		var epic Epic
		if err := yaml.Unmarshal(data, &epic); err != nil {
			return nil, fmt.Errorf("failed to parse epic %s: %w", entry.Name(), err)
		}

		// Store filename without extension for reference
		epic.filename = strings.TrimSuffix(entry.Name(), ".yaml")
		epics = append(epics, epic)
	}

	return epics, nil
}

func loadEpic(basePath, epicName string) (*Epic, error) {
	epicPath := filepath.Join(basePath, ".accelyst", "epics", epicName+".yaml")

	data, err := os.ReadFile(epicPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read epic: %w", err)
	}

	if err := validateYAML(data, epicSchema); err != nil {
		return nil, fmt.Errorf("epic validation: %w", err)
	}

	var epic Epic
	if err := yaml.Unmarshal(data, &epic); err != nil {
		return nil, fmt.Errorf("failed to parse epic: %w", err)
	}

	epic.filename = epicName
	return &epic, nil
}

// Artifact functions

func loadArtifact(basePath, artifactID string) (string, error) {
	artifactPath := filepath.Join(basePath, ".accelyst", "artifacts", artifactID+".md")

	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return "", fmt.Errorf("artifact %q not found: %w", artifactID, err)
	}

	return string(data), nil
}

func artifactExists(basePath, artifactID string) bool {
	artifactPath := filepath.Join(basePath, ".accelyst", "artifacts", artifactID+".md")
	_, err := os.Stat(artifactPath)
	return err == nil
}

// Validation functions

func validateEpicReferences(epic *Epic, registry *Registry, basePath string) error {
	// Build project parts lookup
	projectParts := make(map[string]bool)
	for _, part := range registry.ProjectParts {
		projectParts[part.Name] = true
	}

	// Build artifacts lookup (from registry)
	artifacts := make(map[string]bool)
	for _, artifact := range registry.Artifacts {
		artifacts[artifact.ID] = true
	}

	// Track produces within this epic
	produces := make(map[string]string) // artifact -> epic/milestone/step

	// First pass: collect all produces
	for _, milestone := range epic.Milestones {
		for _, step := range milestone.Steps {
			if step.Produces != "" {
				key := fmt.Sprintf("%s/%s/%s", epic.filename, milestone.ID, step.ID)
				if existing, ok := produces[step.Produces]; ok {
					return fmt.Errorf("artifact %q produced by both %s and %s", step.Produces, existing, key)
				}
				if artifacts[step.Produces] {
					return fmt.Errorf("artifact %q already exists in registry", step.Produces)
				}
				produces[step.Produces] = key
			}
		}
	}

	// Second pass: validate references
	for _, milestone := range epic.Milestones {
		for _, step := range milestone.Steps {
			// Validate project parts
			for _, partName := range step.ProjectParts {
				if !projectParts[partName] {
					return fmt.Errorf("step %s/%s references unknown project part %q", milestone.ID, step.ID, partName)
				}
			}

			// Validate requires
			for _, artifactID := range step.Requires {
				inRegistry := artifacts[artifactID]
				inProduces := produces[artifactID] != ""
				fileExists := artifactExists(basePath, artifactID)

				if !inRegistry && !inProduces && !fileExists {
					return fmt.Errorf("step %s/%s requires unknown artifact %q", milestone.ID, step.ID, artifactID)
				}
			}
		}
	}

	return nil
}

// Topological sort functions

func topologicalSort(steps []Step) ([]Step, error) {
	stepByID := make(map[string]Step)
	inDegree := make(map[string]int)
	adjacency := make(map[string][]string)

	for _, step := range steps {
		stepByID[step.ID] = step
		inDegree[step.ID] = 0
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

	var queue []string
	for _, step := range steps {
		if inDegree[step.ID] == 0 {
			queue = append(queue, step.ID)
		}
	}

	sort.Strings(queue)

	var sorted []Step
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, stepByID[current])

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

	if len(sorted) != len(steps) {
		return nil, fmt.Errorf("cycle detected in step dependencies")
	}

	return sorted, nil
}

func topologicalSortMilestones(milestones []Milestone) ([]Milestone, error) {
	milestoneByID := make(map[string]Milestone)
	inDegree := make(map[string]int)
	adjacency := make(map[string][]string)

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

	var queue []string
	for _, m := range milestones {
		if inDegree[m.ID] == 0 {
			queue = append(queue, m.ID)
		}
	}

	sort.Strings(queue)

	var sorted []Milestone
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, milestoneByID[current])

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

	if len(sorted) != len(milestones) {
		return nil, fmt.Errorf("cycle detected in milestone dependencies")
	}

	return sorted, nil
}

// Prompt generation functions

func generatePromptFragment(step Step, projectParts map[string]ProjectPart, basePath string) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("%s:", step.ID))

	// Steps with no dependencies can run in parallel via subagents
	if len(step.DependsOn) == 0 {
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

	// Add modifies clause
	if len(step.Modifies) > 0 {
		parts = append(parts, fmt.Sprintf("[modifies: %s]", strings.Join(step.Modifies, ", ")))
	}

	return strings.Join(parts, " ")
}

func generateRequiresContext(step Step, basePath string) string {
	if len(step.Requires) == 0 {
		return ""
	}

	var sections []string
	for _, artifactID := range step.Requires {
		content, err := loadArtifact(basePath, artifactID)
		if err != nil {
			sections = append(sections, fmt.Sprintf("### Context: %s\n\n[Artifact not found: %s]\n", artifactID, err))
		} else {
			sections = append(sections, fmt.Sprintf("### Context: %s\n\n%s\n", artifactID, content))
		}
	}

	return strings.Join(sections, "\n")
}

func generateProducesSuffix(step Step) string {
	if step.Produces == "" {
		return ""
	}

	return fmt.Sprintf("\n\n---\nSave analysis to: .accelyst/artifacts/%s.md\nInclude: component locations, state management, key interfaces, extension points", step.Produces)
}

func processMilestone(milestone Milestone, projectParts map[string]ProjectPart, basePath string) (string, error) {
	sortedSteps, err := topologicalSort(milestone.Steps)
	if err != nil {
		return "", fmt.Errorf("milestone %q: %w", milestone.Name, err)
	}

	var output []string
	for _, step := range sortedSteps {
		var stepOutput []string

		// Add requires context before the step
		requiresContext := generateRequiresContext(step, basePath)
		if requiresContext != "" {
			stepOutput = append(stepOutput, requiresContext)
			stepOutput = append(stepOutput, "---\n")
		}

		// Add the main prompt fragment
		fragment := generatePromptFragment(step, projectParts, basePath)
		stepOutput = append(stepOutput, fragment)

		// Add produces suffix
		producesSuffix := generateProducesSuffix(step)
		if producesSuffix != "" {
			stepOutput = append(stepOutput, producesSuffix)
		}

		output = append(output, strings.Join(stepOutput, ""))
	}

	return strings.Join(output, "\n\n"), nil
}

// Command implementations

func cmdInit(basePath string) error {
	// Convert to absolute path
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	accelystDir := filepath.Join(absBasePath, ".accelyst")
	epicsDir := filepath.Join(accelystDir, "epics")
	artifactsDir := filepath.Join(accelystDir, "artifacts")
	registryPath := filepath.Join(accelystDir, "registry.yaml")

	// Check if already initialized
	if _, err := os.Stat(registryPath); err == nil {
		return fmt.Errorf("already initialized: %s exists", registryPath)
	}

	// Create directories
	if err := os.MkdirAll(epicsDir, 0755); err != nil {
		return fmt.Errorf("failed to create epics directory: %w", err)
	}
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Create initial registry
	registry := Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{
				Name:         "app",
				DirectoryAbs: absBasePath,
				Description:  "Main application",
			},
			{
				Name:         "test",
				DirectoryAbs: filepath.Join(absBasePath, "tests"),
				Description:  "Test suites",
			},
		},
		Artifacts: []Artifact{},
	}

	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	fmt.Printf("Initialized accelyst in %s\n", accelystDir)
	fmt.Printf("  - Created %s\n", registryPath)
	fmt.Printf("  - Created %s/\n", epicsDir)
	fmt.Printf("  - Created %s/\n", artifactsDir)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .accelyst/registry.yaml to configure project parts")
	fmt.Println("  2. Create epic files in .accelyst/epics/")

	return nil
}

func cmdMigrate(basePath, configPath string) error {
	// Read legacy config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var legacy LegacyConfig
	if err := yaml.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Create directories
	accelystDir := filepath.Join(basePath, ".accelyst")
	epicsDir := filepath.Join(accelystDir, "epics")
	artifactsDir := filepath.Join(accelystDir, "artifacts")

	if err := os.MkdirAll(epicsDir, 0755); err != nil {
		return fmt.Errorf("failed to create epics directory: %w", err)
	}
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Create registry from project parts
	registry := Registry{
		Version:      1,
		ProjectParts: legacy.ProjectParts,
		Artifacts:    []Artifact{},
	}

	registryData, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	registryPath := filepath.Join(accelystDir, "registry.yaml")
	if err := os.WriteFile(registryPath, registryData, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	// Create epic from milestones
	epic := Epic{
		Name:        "Migrated",
		Description: "Milestones migrated from legacy config",
		Priority:    "ready",
		Milestones:  legacy.Milestones,
	}

	epicData, err := yaml.Marshal(epic)
	if err != nil {
		return fmt.Errorf("failed to marshal epic: %w", err)
	}

	epicPath := filepath.Join(epicsDir, "default.yaml")
	if err := os.WriteFile(epicPath, epicData, 0644); err != nil {
		return fmt.Errorf("failed to write epic: %w", err)
	}

	// Backup original config
	backupPath := configPath + ".bak"
	if err := os.Rename(configPath, backupPath); err != nil {
		fmt.Printf("Warning: failed to backup config: %v\n", err)
	} else {
		fmt.Printf("Backed up original config to %s\n", backupPath)
	}

	fmt.Printf("Migration complete:\n")
	fmt.Printf("  - Created %s\n", registryPath)
	fmt.Printf("  - Created %s\n", epicPath)

	return nil
}

func cmdEpicsList(basePath string) error {
	epics, err := loadEpics(basePath)
	if err != nil {
		return err
	}

	if len(epics) == 0 {
		fmt.Println("No epics found in .accelyst/epics/")
		return nil
	}

	for _, epic := range epics {
		priority := epic.Priority
		if priority == "" {
			priority = "none"
		}
		fmt.Printf("%s (%s)\n", epic.filename, priority)
		fmt.Printf("  %s\n", epic.Name)
		if epic.Description != "" {
			fmt.Printf("  %s\n", epic.Description)
		}
		fmt.Printf("  Milestones: %d\n", len(epic.Milestones))
		fmt.Println()
	}

	return nil
}

func cmdRegistryParts(basePath string) error {
	registry, err := loadRegistry(basePath)
	if err != nil {
		return err
	}

	if len(registry.ProjectParts) == 0 {
		fmt.Println("No project parts defined")
		return nil
	}

	for _, part := range registry.ProjectParts {
		fmt.Printf("%s\n", part.Name)
		fmt.Printf("  Directory: %s\n", part.DirectoryAbs)
		if part.DocumentationAbs != "" {
			fmt.Printf("  Documentation: %s\n", part.DocumentationAbs)
		}
		if part.Description != "" {
			fmt.Printf("  Description: %s\n", part.Description)
		}
		fmt.Println()
	}

	return nil
}

func cmdRegistryArtifacts(basePath string) error {
	registry, err := loadRegistry(basePath)
	if err != nil {
		return err
	}

	if len(registry.Artifacts) == 0 {
		fmt.Println("No artifacts registered")
		return nil
	}

	for _, artifact := range registry.Artifacts {
		fmt.Printf("%s\n", artifact.ID)
		fmt.Printf("  Description: %s\n", artifact.Description)
		fmt.Printf("  Produced by: %s\n", artifact.ProducedBy)
		fmt.Printf("  Created: %s\n", artifact.CreatedAt)

		// Check if file exists
		if artifactExists(basePath, artifact.ID) {
			fmt.Printf("  File: .accelyst/artifacts/%s.md (exists)\n", artifact.ID)
		} else {
			fmt.Printf("  File: .accelyst/artifacts/%s.md (MISSING)\n", artifact.ID)
		}
		fmt.Println()
	}

	return nil
}

func cmdRun(basePath, target string) error {
	// Parse target: epic/milestone or epic/milestone/step
	parts := strings.Split(target, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid target: expected epic/milestone or epic/milestone/step")
	}

	epicName := parts[0]
	milestoneID := parts[1]
	var stepID string
	if len(parts) >= 3 {
		stepID = parts[2]
	}

	// Load registry
	registry, err := loadRegistry(basePath)
	if err != nil {
		return err
	}

	// Load epic
	epic, err := loadEpic(basePath, epicName)
	if err != nil {
		return err
	}

	// Validate references
	if err := validateEpicReferences(epic, registry, basePath); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Build project parts lookup
	projectParts := make(map[string]ProjectPart)
	for _, part := range registry.ProjectParts {
		projectParts[part.Name] = part
	}

	// Find milestone
	var milestone *Milestone
	for i := range epic.Milestones {
		if epic.Milestones[i].ID == milestoneID {
			milestone = &epic.Milestones[i]
			break
		}
	}
	if milestone == nil {
		return fmt.Errorf("milestone %q not found in epic %q", milestoneID, epicName)
	}

	// If step ID provided, filter to just that step
	if stepID != "" {
		var step *Step
		for _, s := range milestone.Steps {
			if s.ID == stepID {
				step = &s
				break
			}
		}
		if step == nil {
			return fmt.Errorf("step %q not found in milestone %q", stepID, milestoneID)
		}

		// Output just this step
		fmt.Printf("# Step: %s\n\n", step.ID)

		requiresContext := generateRequiresContext(*step, basePath)
		if requiresContext != "" {
			fmt.Print(requiresContext)
			fmt.Println("---")
		}

		fragment := generatePromptFragment(*step, projectParts, basePath)
		fmt.Println(fragment)

		producesSuffix := generateProducesSuffix(*step)
		if producesSuffix != "" {
			fmt.Println(producesSuffix)
		}

		return nil
	}

	// Output full milestone
	header := fmt.Sprintf("# Milestone: %s", milestone.Name)
	if len(milestone.DependsOn) > 0 {
		header += fmt.Sprintf(" (depends on: %s)", strings.Join(milestone.DependsOn, ", "))
	}
	fmt.Println(header)
	fmt.Println()

	prompt, err := processMilestone(*milestone, projectParts, basePath)
	if err != nil {
		return err
	}

	fmt.Println(prompt)

	return nil
}

func cmdRegistryAddArtifact(basePath, artifactID, description, producedBy string) error {
	registry, err := loadRegistry(basePath)
	if err != nil {
		return err
	}

	// Check if artifact already exists
	for _, a := range registry.Artifacts {
		if a.ID == artifactID {
			return fmt.Errorf("artifact %q already exists", artifactID)
		}
	}

	// Check if artifact file exists
	if !artifactExists(basePath, artifactID) {
		return fmt.Errorf("artifact file not found: .accelyst/artifacts/%s.md", artifactID)
	}

	// Add artifact
	artifact := Artifact{
		ID:          artifactID,
		Description: description,
		ProducedBy:  producedBy,
		CreatedAt:   time.Now().Format("2006-01-02"),
	}
	registry.Artifacts = append(registry.Artifacts, artifact)

	// Save registry
	if err := saveRegistry(basePath, registry); err != nil {
		return err
	}

	fmt.Printf("Added artifact %q to registry\n", artifactID)
	return nil
}

func main() {
	// Define flags
	basePath := flag.String("path", ".", "base path for project")
	configPath := flag.String("config", "", "legacy config path (for migrate command)")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: accelyst [flags] <command> [args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  init                      Initialize accelyst in current directory")
		fmt.Println("  migrate                   Migrate from legacy accelyst.yaml")
		fmt.Println("  epics list                List all epics")
		fmt.Println("  registry parts            List project parts")
		fmt.Println("  registry artifacts        List registered artifacts")
		fmt.Println("  registry add-artifact     Register an artifact")
		fmt.Println("  run <epic/milestone>      Run a milestone")
		fmt.Println("  run <epic/milestone/step> Run a specific step")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "init":
		err = cmdInit(*basePath)

	case "migrate":
		cfg := *configPath
		if cfg == "" {
			cfg = filepath.Join(*basePath, "accelyst.yaml")
		}
		err = cmdMigrate(*basePath, cfg)

	case "epics":
		if len(args) < 2 {
			err = fmt.Errorf("usage: accelyst epics <list>")
		} else if args[1] == "list" {
			err = cmdEpicsList(*basePath)
		} else {
			err = fmt.Errorf("unknown epics command: %s", args[1])
		}

	case "registry":
		if len(args) < 2 {
			err = fmt.Errorf("usage: accelyst registry <parts|artifacts|add-artifact>")
		} else {
			switch args[1] {
			case "parts":
				err = cmdRegistryParts(*basePath)
			case "artifacts":
				err = cmdRegistryArtifacts(*basePath)
			case "add-artifact":
				if len(args) < 5 {
					err = fmt.Errorf("usage: accelyst registry add-artifact <id> <description> <producedBy>")
				} else {
					err = cmdRegistryAddArtifact(*basePath, args[2], args[3], args[4])
				}
			default:
				err = fmt.Errorf("unknown registry command: %s", args[1])
			}
		}

	case "run":
		if len(args) < 2 {
			err = fmt.Errorf("usage: accelyst run <epic/milestone> or <epic/milestone/step>")
		} else {
			err = cmdRun(*basePath, args[1])
		}

	default:
		err = fmt.Errorf("unknown command: %s", args[0])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
