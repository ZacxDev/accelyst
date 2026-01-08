package main

import (
	"strings"
	"testing"
)

// ============================================
// Schema Validation Tests
// ============================================

func TestValidateConfig_Valid(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - id: test
    name: "Test Milestone"
    steps:
      - id: step1
        instruction: "do something"
`
	if err := validateConfig([]byte(yaml)); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestValidateConfig_MissingMilestoneID(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - name: "Missing ID"
    steps:
      - id: step1
        instruction: "do something"
`
	err := validateConfig([]byte(yaml))
	if err == nil {
		t.Error("expected error for missing milestone id")
	}
	if !strings.Contains(err.Error(), "id is required") {
		t.Errorf("expected 'id is required' error, got: %v", err)
	}
}

func TestValidateConfig_MissingStepInstruction(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - id: test
    name: "Test"
    steps:
      - id: step1
`
	err := validateConfig([]byte(yaml))
	if err == nil {
		t.Error("expected error for missing instruction")
	}
	if !strings.Contains(err.Error(), "instruction is required") {
		t.Errorf("expected 'instruction is required' error, got: %v", err)
	}
}

func TestValidateConfig_EmptySteps(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - id: test
    name: "Test"
    steps: []
`
	err := validateConfig([]byte(yaml))
	if err == nil {
		t.Error("expected error for empty steps")
	}
	if !strings.Contains(err.Error(), "at least 1 items") {
		t.Errorf("expected 'at least 1 items' error, got: %v", err)
	}
}

func TestValidateConfig_InvalidTier(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - id: test
    name: "Test"
    steps:
      - id: step1
        instruction: "do something"
        tier: invalid
`
	err := validateConfig([]byte(yaml))
	if err == nil {
		t.Error("expected error for invalid tier")
	}
	if !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("expected 'must be one of' error, got: %v", err)
	}
}

func TestValidateConfig_ValidTiers(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - id: test
    name: "Test"
    steps:
      - id: step1
        instruction: "ai task"
        tier: ai
      - id: step2
        instruction: "human task"
        tier: human
`
	if err := validateConfig([]byte(yaml)); err != nil {
		t.Errorf("expected valid config with tiers, got error: %v", err)
	}
}

func TestValidateConfig_AcceptanceCriteria(t *testing.T) {
	yaml := `
projectParts:
  - name: client
    directoryAbs: /path/to/client

milestones:
  - id: test
    name: "Test"
    steps:
      - id: step1
        instruction: "do something"
        acceptanceCriteria:
          - "condition 1"
          - "condition 2"
`
	if err := validateConfig([]byte(yaml)); err != nil {
		t.Errorf("expected valid config with acceptanceCriteria, got error: %v", err)
	}
}

// ============================================
// Topological Sort Tests - Steps
// ============================================

func TestTopologicalSort_NoDependencies(t *testing.T) {
	steps := []Step{
		{ID: "a", Instruction: "do a"},
		{ID: "b", Instruction: "do b"},
		{ID: "c", Instruction: "do c"},
	}

	sorted, err := topologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("expected 3 steps, got %d", len(sorted))
	}
}

func TestTopologicalSort_LinearDependencies(t *testing.T) {
	steps := []Step{
		{ID: "c", Instruction: "do c", DependsOn: []string{"b"}},
		{ID: "a", Instruction: "do a"},
		{ID: "b", Instruction: "do b", DependsOn: []string{"a"}},
	}

	sorted, err := topologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be a -> b -> c
	if sorted[0].ID != "a" || sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("expected order a,b,c got %s,%s,%s", sorted[0].ID, sorted[1].ID, sorted[2].ID)
	}
}

func TestTopologicalSort_DiamondDependencies(t *testing.T) {
	steps := []Step{
		{ID: "d", Instruction: "do d", DependsOn: []string{"b", "c"}},
		{ID: "b", Instruction: "do b", DependsOn: []string{"a"}},
		{ID: "c", Instruction: "do c", DependsOn: []string{"a"}},
		{ID: "a", Instruction: "do a"},
	}

	sorted, err := topologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// a must come first, d must come last
	if sorted[0].ID != "a" {
		t.Errorf("expected 'a' first, got %s", sorted[0].ID)
	}
	if sorted[3].ID != "d" {
		t.Errorf("expected 'd' last, got %s", sorted[3].ID)
	}
}

func TestTopologicalSort_CycleDetection(t *testing.T) {
	steps := []Step{
		{ID: "a", Instruction: "do a", DependsOn: []string{"b"}},
		{ID: "b", Instruction: "do b", DependsOn: []string{"a"}},
	}

	_, err := topologicalSort(steps)
	if err == nil {
		t.Error("expected cycle detection error")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected 'cycle detected' error, got: %v", err)
	}
}

func TestTopologicalSort_InvalidDependency(t *testing.T) {
	steps := []Step{
		{ID: "a", Instruction: "do a", DependsOn: []string{"nonexistent"}},
	}

	_, err := topologicalSort(steps)
	if err == nil {
		t.Error("expected invalid dependency error")
	}
	if !strings.Contains(err.Error(), "non-existent step") {
		t.Errorf("expected 'non-existent step' error, got: %v", err)
	}
}

// ============================================
// Topological Sort Tests - Milestones
// ============================================

func TestTopologicalSortMilestones_NoDependencies(t *testing.T) {
	milestones := []Milestone{
		{ID: "a", Name: "A", Steps: []Step{{ID: "s1", Instruction: "x"}}},
		{ID: "b", Name: "B", Steps: []Step{{ID: "s1", Instruction: "x"}}},
	}

	sorted, err := topologicalSortMilestones(milestones)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sorted) != 2 {
		t.Errorf("expected 2 milestones, got %d", len(sorted))
	}
}

func TestTopologicalSortMilestones_WithDependencies(t *testing.T) {
	milestones := []Milestone{
		{ID: "c", Name: "C", DependsOn: []string{"a", "b"}, Steps: []Step{{ID: "s1", Instruction: "x"}}},
		{ID: "a", Name: "A", Steps: []Step{{ID: "s1", Instruction: "x"}}},
		{ID: "b", Name: "B", Steps: []Step{{ID: "s1", Instruction: "x"}}},
	}

	sorted, err := topologicalSortMilestones(milestones)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// c must come last
	if sorted[2].ID != "c" {
		t.Errorf("expected 'c' last, got %s", sorted[2].ID)
	}
}

func TestTopologicalSortMilestones_CycleDetection(t *testing.T) {
	milestones := []Milestone{
		{ID: "a", Name: "A", DependsOn: []string{"b"}, Steps: []Step{{ID: "s1", Instruction: "x"}}},
		{ID: "b", Name: "B", DependsOn: []string{"a"}, Steps: []Step{{ID: "s1", Instruction: "x"}}},
	}

	_, err := topologicalSortMilestones(milestones)
	if err == nil {
		t.Error("expected cycle detection error")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected 'cycle detected' error, got: %v", err)
	}
}

func TestTopologicalSortMilestones_MissingID(t *testing.T) {
	milestones := []Milestone{
		{Name: "No ID", Steps: []Step{{ID: "s1", Instruction: "x"}}},
	}

	_, err := topologicalSortMilestones(milestones)
	if err == nil {
		t.Error("expected missing id error")
	}
	if !strings.Contains(err.Error(), "missing required 'id' field") {
		t.Errorf("expected 'missing required id field' error, got: %v", err)
	}
}

func TestTopologicalSortMilestones_DuplicateID(t *testing.T) {
	milestones := []Milestone{
		{ID: "same", Name: "A", Steps: []Step{{ID: "s1", Instruction: "x"}}},
		{ID: "same", Name: "B", Steps: []Step{{ID: "s1", Instruction: "x"}}},
	}

	_, err := topologicalSortMilestones(milestones)
	if err == nil {
		t.Error("expected duplicate id error")
	}
	if !strings.Contains(err.Error(), "duplicate milestone id") {
		t.Errorf("expected 'duplicate milestone id' error, got: %v", err)
	}
}

// ============================================
// Prompt Generation Tests
// ============================================

func TestGeneratePromptFragment_Basic(t *testing.T) {
	step := Step{
		ID:          "test",
		Instruction: "do something",
	}

	result := generatePromptFragment(step, nil)
	expected := "test: use a subagent to do something"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGeneratePromptFragment_WithDependencies(t *testing.T) {
	step := Step{
		ID:          "test",
		Instruction: "do something",
		DependsOn:   []string{"step1", "step2"},
	}

	result := generatePromptFragment(step, nil)

	if strings.Contains(result, "use a subagent to") {
		t.Error("step with dependencies should not have subagent prefix")
	}
	if !strings.Contains(result, "read results from steps step1, step2") {
		t.Errorf("expected dependency clause, got: %s", result)
	}
}

func TestGeneratePromptFragment_WithProjectParts(t *testing.T) {
	step := Step{
		ID:           "test",
		Instruction:  "do something",
		ProjectParts: []string{"client"},
	}

	projectParts := map[string]ProjectPart{
		"client": {Name: "client", DirectoryAbs: "/path/to/client"},
	}

	result := generatePromptFragment(step, projectParts)

	if !strings.Contains(result, "Analyze client (/path/to/client)") {
		t.Errorf("expected project part analysis, got: %s", result)
	}
}

func TestGeneratePromptFragment_WithDocumentation(t *testing.T) {
	step := Step{
		ID:           "test",
		Instruction:  "do something",
		ProjectParts: []string{"client"},
	}

	projectParts := map[string]ProjectPart{
		"client": {Name: "client", DirectoryAbs: "/path/to/client", DocumentationAbs: "/path/to/docs.md"},
	}

	result := generatePromptFragment(step, projectParts)

	if !strings.Contains(result, "(docs: /path/to/docs.md)") {
		t.Errorf("expected documentation path, got: %s", result)
	}
}

func TestGeneratePromptFragment_WithAcceptanceCriteria(t *testing.T) {
	step := Step{
		ID:                 "test",
		Instruction:        "do something",
		AcceptanceCriteria: []string{"condition 1", "condition 2"},
	}

	result := generatePromptFragment(step, nil)

	if !strings.Contains(result, "[done when: condition 1; condition 2]") {
		t.Errorf("expected acceptance criteria, got: %s", result)
	}
}

func TestGeneratePromptFragment_HumanTier(t *testing.T) {
	step := Step{
		ID:          "test",
		Instruction: "configure oauth",
		Tier:        "human",
	}

	result := generatePromptFragment(step, nil)

	if !strings.Contains(result, "[HUMAN]") {
		t.Errorf("expected [HUMAN] prefix, got: %s", result)
	}
	if strings.Contains(result, "use a subagent to") {
		t.Error("human tier should not have subagent prefix")
	}
}

func TestGeneratePromptFragment_AITier(t *testing.T) {
	step := Step{
		ID:          "test",
		Instruction: "do something",
		Tier:        "ai",
	}

	result := generatePromptFragment(step, nil)

	if strings.Contains(result, "[HUMAN]") {
		t.Error("ai tier should not have [HUMAN] prefix")
	}
	if !strings.Contains(result, "use a subagent to") {
		t.Errorf("ai tier with no deps should have subagent prefix, got: %s", result)
	}
}

func TestGeneratePromptFragment_HumanTierWithDependencies(t *testing.T) {
	step := Step{
		ID:          "test",
		Instruction: "configure oauth",
		Tier:        "human",
		DependsOn:   []string{"prev"},
	}

	result := generatePromptFragment(step, nil)

	if !strings.Contains(result, "[HUMAN]") {
		t.Errorf("expected [HUMAN] prefix, got: %s", result)
	}
	if !strings.Contains(result, "read results from steps prev") {
		t.Errorf("expected dependency clause, got: %s", result)
	}
}

// ============================================
// Process Milestone Tests
// ============================================

func TestProcessMilestone_Basic(t *testing.T) {
	milestone := Milestone{
		ID:   "test",
		Name: "Test",
		Steps: []Step{
			{ID: "a", Instruction: "do a"},
			{ID: "b", Instruction: "do b", DependsOn: []string{"a"}},
		},
	}

	result, err := processMilestone(milestone, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// First line should be step a (no dependencies)
	if !strings.HasPrefix(lines[0], "a:") {
		t.Errorf("expected first line to start with 'a:', got: %s", lines[0])
	}
}

// ============================================
// YAML to JSON Conversion Tests
// ============================================

func TestConvertYAMLToJSON_MapConversion(t *testing.T) {
	input := map[interface{}]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	result := convertYAMLToJSON(input)

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}

	if m["key1"] != "value1" {
		t.Errorf("expected 'value1', got %v", m["key1"])
	}
}

func TestConvertYAMLToJSON_NestedMap(t *testing.T) {
	input := map[interface{}]interface{}{
		"outer": map[interface{}]interface{}{
			"inner": "value",
		},
	}

	result := convertYAMLToJSON(input)

	m := result.(map[string]interface{})
	inner := m["outer"].(map[string]interface{})

	if inner["inner"] != "value" {
		t.Errorf("expected 'value', got %v", inner["inner"])
	}
}

func TestConvertYAMLToJSON_Array(t *testing.T) {
	input := []interface{}{
		map[interface{}]interface{}{"key": "value"},
	}

	result := convertYAMLToJSON(input)

	arr := result.([]interface{})
	m := arr[0].(map[string]interface{})

	if m["key"] != "value" {
		t.Errorf("expected 'value', got %v", m["key"])
	}
}
