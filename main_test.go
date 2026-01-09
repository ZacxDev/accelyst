package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================
// Test Helpers
// ============================================

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "accelyst-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

func cleanupTestDir(t *testing.T, dir string) {
	t.Helper()
	os.RemoveAll(dir)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// ============================================
// Registry Tests
// ============================================

func TestLoadRegistry_Valid(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	registry := `version: 1
projectParts:
  - name: client
    directoryAbs: /path/to/client
  - name: api
    directoryAbs: /path/to/api
artifacts:
  - id: auth_flow
    description: "Authentication flow analysis"
    producedBy: "epic/milestone/step"
    createdAt: "2024-01-08"
`
	writeFile(t, filepath.Join(dir, ".accelyst", "registry.yaml"), registry)

	reg, err := loadRegistry(dir)
	if err != nil {
		t.Fatalf("loadRegistry failed: %v", err)
	}

	if reg.Version != 1 {
		t.Errorf("Version = %d, want 1", reg.Version)
	}
	if len(reg.ProjectParts) != 2 {
		t.Errorf("len(ProjectParts) = %d, want 2", len(reg.ProjectParts))
	}
	if len(reg.Artifacts) != 1 {
		t.Errorf("len(Artifacts) = %d, want 1", len(reg.Artifacts))
	}
	if reg.Artifacts[0].ID != "auth_flow" {
		t.Errorf("Artifacts[0].ID = %q, want %q", reg.Artifacts[0].ID, "auth_flow")
	}
}

func TestLoadRegistry_MissingVersion(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	registry := `projectParts:
  - name: client
    directoryAbs: /path/to/client
`
	writeFile(t, filepath.Join(dir, ".accelyst", "registry.yaml"), registry)

	_, err := loadRegistry(dir)
	if err == nil {
		t.Error("Expected error for missing version, got nil")
	}
}

func TestLoadRegistry_MissingProjectParts(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	registry := `version: 1
`
	writeFile(t, filepath.Join(dir, ".accelyst", "registry.yaml"), registry)

	_, err := loadRegistry(dir)
	if err == nil {
		t.Error("Expected error for missing projectParts, got nil")
	}
}

// ============================================
// Epic Tests
// ============================================

func TestLoadEpic_Valid(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	epic := `name: "Test Epic"
description: "A test epic"
priority: ready
milestones:
  - id: milestone_1
    name: "Milestone One"
    steps:
      - id: step_1
        instruction: "do something"
`
	writeFile(t, filepath.Join(dir, ".accelyst", "epics", "test.yaml"), epic)

	e, err := loadEpic(dir, "test")
	if err != nil {
		t.Fatalf("loadEpic failed: %v", err)
	}

	if e.Name != "Test Epic" {
		t.Errorf("Name = %q, want %q", e.Name, "Test Epic")
	}
	if e.Priority != "ready" {
		t.Errorf("Priority = %q, want %q", e.Priority, "ready")
	}
	if len(e.Milestones) != 1 {
		t.Errorf("len(Milestones) = %d, want 1", len(e.Milestones))
	}
}

func TestLoadEpic_InvalidPriority(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	epic := `name: "Test Epic"
priority: invalid_priority
milestones:
  - id: m1
    name: "M1"
    steps:
      - id: s1
        instruction: "do"
`
	writeFile(t, filepath.Join(dir, ".accelyst", "epics", "test.yaml"), epic)

	_, err := loadEpic(dir, "test")
	if err == nil {
		t.Error("Expected error for invalid priority, got nil")
	}
}

func TestLoadEpics_Multiple(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	epic1 := `name: "Epic One"
milestones:
  - id: m1
    name: "M1"
    steps:
      - id: s1
        instruction: "do"
`
	epic2 := `name: "Epic Two"
milestones:
  - id: m2
    name: "M2"
    steps:
      - id: s2
        instruction: "do"
`
	writeFile(t, filepath.Join(dir, ".accelyst", "epics", "one.yaml"), epic1)
	writeFile(t, filepath.Join(dir, ".accelyst", "epics", "two.yaml"), epic2)

	epics, err := loadEpics(dir)
	if err != nil {
		t.Fatalf("loadEpics failed: %v", err)
	}

	if len(epics) != 2 {
		t.Errorf("len(epics) = %d, want 2", len(epics))
	}
}

// ============================================
// Validation Tests
// ============================================

func TestValidateEpicReferences_ValidProjectParts(t *testing.T) {
	registry := &Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{Name: "client", DirectoryAbs: "/path/to/client"},
			{Name: "api", DirectoryAbs: "/path/to/api"},
		},
	}

	epic := &Epic{
		filename: "test",
		Milestones: []Milestone{
			{
				ID:   "m1",
				Name: "M1",
				Steps: []Step{
					{ID: "s1", Instruction: "do", ProjectParts: []string{"client"}},
				},
			},
		},
	}

	err := validateEpicReferences(epic, registry, "/tmp")
	if err != nil {
		t.Errorf("validateEpicReferences failed: %v", err)
	}
}

func TestValidateEpicReferences_InvalidProjectPart(t *testing.T) {
	registry := &Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{Name: "client", DirectoryAbs: "/path/to/client"},
		},
	}

	epic := &Epic{
		filename: "test",
		Milestones: []Milestone{
			{
				ID:   "m1",
				Name: "M1",
				Steps: []Step{
					{ID: "s1", Instruction: "do", ProjectParts: []string{"unknown"}},
				},
			},
		},
	}

	err := validateEpicReferences(epic, registry, "/tmp")
	if err == nil {
		t.Error("Expected error for unknown project part, got nil")
	}
	if !strings.Contains(err.Error(), "unknown project part") {
		t.Errorf("Error message %q should contain 'unknown project part'", err.Error())
	}
}

func TestValidateEpicReferences_DuplicateProduces(t *testing.T) {
	registry := &Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{Name: "client", DirectoryAbs: "/path/to/client"},
		},
	}

	epic := &Epic{
		filename: "test",
		Milestones: []Milestone{
			{
				ID:   "m1",
				Name: "M1",
				Steps: []Step{
					{ID: "s1", Instruction: "do", Produces: "artifact_a"},
					{ID: "s2", Instruction: "do", Produces: "artifact_a"},
				},
			},
		},
	}

	err := validateEpicReferences(epic, registry, "/tmp")
	if err == nil {
		t.Error("Expected error for duplicate produces, got nil")
	}
	if !strings.Contains(err.Error(), "produced by both") {
		t.Errorf("Error message %q should contain 'produced by both'", err.Error())
	}
}

func TestValidateEpicReferences_ProducesExistsInRegistry(t *testing.T) {
	registry := &Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{Name: "client", DirectoryAbs: "/path/to/client"},
		},
		Artifacts: []Artifact{
			{ID: "existing_artifact", Description: "exists", ProducedBy: "old/m/s", CreatedAt: "2024-01-01"},
		},
	}

	epic := &Epic{
		filename: "test",
		Milestones: []Milestone{
			{
				ID:   "m1",
				Name: "M1",
				Steps: []Step{
					{ID: "s1", Instruction: "do", Produces: "existing_artifact"},
				},
			},
		},
	}

	err := validateEpicReferences(epic, registry, "/tmp")
	if err == nil {
		t.Error("Expected error for artifact already in registry, got nil")
	}
	if !strings.Contains(err.Error(), "already exists in registry") {
		t.Errorf("Error message %q should contain 'already exists in registry'", err.Error())
	}
}

func TestValidateEpicReferences_RequiresFromProduces(t *testing.T) {
	registry := &Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{Name: "client", DirectoryAbs: "/path/to/client"},
		},
	}

	epic := &Epic{
		filename: "test",
		Milestones: []Milestone{
			{
				ID:   "m1",
				Name: "M1",
				Steps: []Step{
					{ID: "s1", Instruction: "analyze", Produces: "analysis"},
					{ID: "s2", Instruction: "use", Requires: []string{"analysis"}, DependsOn: []string{"s1"}},
				},
			},
		},
	}

	err := validateEpicReferences(epic, registry, "/tmp")
	if err != nil {
		t.Errorf("validateEpicReferences failed: %v", err)
	}
}

func TestValidateEpicReferences_RequiresUnknown(t *testing.T) {
	registry := &Registry{
		Version: 1,
		ProjectParts: []ProjectPart{
			{Name: "client", DirectoryAbs: "/path/to/client"},
		},
	}

	epic := &Epic{
		filename: "test",
		Milestones: []Milestone{
			{
				ID:   "m1",
				Name: "M1",
				Steps: []Step{
					{ID: "s1", Instruction: "use", Requires: []string{"nonexistent"}},
				},
			},
		},
	}

	err := validateEpicReferences(epic, registry, "/tmp")
	if err == nil {
		t.Error("Expected error for unknown requires, got nil")
	}
	if !strings.Contains(err.Error(), "unknown artifact") {
		t.Errorf("Error message %q should contain 'unknown artifact'", err.Error())
	}
}

// ============================================
// Topological Sort Tests - Steps
// ============================================

func TestTopologicalSort_NoDependencies(t *testing.T) {
	steps := []Step{
		{ID: "c", Instruction: "do c"},
		{ID: "a", Instruction: "do a"},
		{ID: "b", Instruction: "do b"},
	}

	sorted, err := topologicalSort(steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("expected 3 steps, got %d", len(sorted))
	}

	// All have no dependencies, should be sorted alphabetically
	if sorted[0].ID != "a" || sorted[1].ID != "b" || sorted[2].ID != "c" {
		t.Errorf("Expected alphabetical order, got: %s, %s, %s", sorted[0].ID, sorted[1].ID, sorted[2].ID)
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

	result := generatePromptFragment(step, nil, "/tmp")
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

	result := generatePromptFragment(step, nil, "/tmp")

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

	result := generatePromptFragment(step, projectParts, "/tmp")

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

	result := generatePromptFragment(step, projectParts, "/tmp")

	if !strings.Contains(result, "(docs: /path/to/docs.md)") {
		t.Errorf("expected documentation path, got: %s", result)
	}
}

func TestGenerateProducesSuffix_NoProduces(t *testing.T) {
	step := Step{
		ID:          "impl",
		Instruction: "implement",
	}

	suffix := generateProducesSuffix(step)

	if suffix != "" {
		t.Errorf("Expected empty suffix for step without produces, got %q", suffix)
	}
}

func TestGenerateProducesSuffix_WithProduces(t *testing.T) {
	step := Step{
		ID:          "analyze",
		Instruction: "analyze",
		Produces:    "analysis_output",
	}

	suffix := generateProducesSuffix(step)

	if !strings.Contains(suffix, "analysis_output.md") {
		t.Errorf("Suffix should contain artifact filename, got %q", suffix)
	}
	if !strings.Contains(suffix, "Save analysis to:") {
		t.Errorf("Suffix should contain save instruction, got %q", suffix)
	}
}

func TestGenerateRequiresContext_NoRequires(t *testing.T) {
	step := Step{
		ID:          "impl",
		Instruction: "implement",
	}

	context := generateRequiresContext(step, "/tmp")

	if context != "" {
		t.Errorf("Expected empty context for step without requires, got %q", context)
	}
}

func TestGenerateRequiresContext_WithRequires(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	// Create artifact file
	artifactContent := "# Analysis\n\nThis is the analysis."
	writeFile(t, filepath.Join(dir, ".accelyst", "artifacts", "my_analysis.md"), artifactContent)

	step := Step{
		ID:          "impl",
		Instruction: "implement",
		Requires:    []string{"my_analysis"},
	}

	context := generateRequiresContext(step, dir)

	if !strings.Contains(context, "### Context: my_analysis") {
		t.Errorf("Context should contain header, got %q", context)
	}
	if !strings.Contains(context, "This is the analysis.") {
		t.Errorf("Context should contain artifact content, got %q", context)
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

	result, err := processMilestone(milestone, nil, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain both steps
	if !strings.Contains(result, "a:") {
		t.Errorf("expected step a, got: %s", result)
	}
	if !strings.Contains(result, "b:") {
		t.Errorf("expected step b, got: %s", result)
	}
}

// ============================================
// Init Command Tests
// ============================================

func TestCmdInit_CreatesStructure(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	err := cmdInit(dir)
	if err != nil {
		t.Fatalf("cmdInit failed: %v", err)
	}

	// Check registry exists
	if _, err := os.Stat(filepath.Join(dir, ".accelyst", "registry.yaml")); os.IsNotExist(err) {
		t.Error("Registry file not created")
	}

	// Check epics directory exists
	if _, err := os.Stat(filepath.Join(dir, ".accelyst", "epics")); os.IsNotExist(err) {
		t.Error("Epics directory not created")
	}

	// Check artifacts directory exists
	if _, err := os.Stat(filepath.Join(dir, ".accelyst", "artifacts")); os.IsNotExist(err) {
		t.Error("Artifacts directory not created")
	}
}

func TestCmdInit_AlreadyInitialized(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	// First init should succeed
	err := cmdInit(dir)
	if err != nil {
		t.Fatalf("First cmdInit failed: %v", err)
	}

	// Second init should fail
	err = cmdInit(dir)
	if err == nil {
		t.Error("Expected error for already initialized, got nil")
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Errorf("Error message %q should contain 'already initialized'", err.Error())
	}
}

// ============================================
// Migrate Command Tests
// ============================================

func TestCmdMigrate_Valid(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	// Create legacy config
	legacyConfig := `projectParts:
  - name: client
    directoryAbs: /path/to/client
  - name: api
    directoryAbs: /path/to/api
milestones:
  - id: feature_1
    name: "Feature One"
    steps:
      - id: step_1
        instruction: "do something"
`
	configPath := filepath.Join(dir, "accelyst.yaml")
	writeFile(t, configPath, legacyConfig)

	err := cmdMigrate(dir, configPath)
	if err != nil {
		t.Fatalf("cmdMigrate failed: %v", err)
	}

	// Check registry was created
	registry, err := loadRegistry(dir)
	if err != nil {
		t.Fatalf("Failed to load migrated registry: %v", err)
	}
	if len(registry.ProjectParts) != 2 {
		t.Errorf("len(ProjectParts) = %d, want 2", len(registry.ProjectParts))
	}

	// Check epic was created
	epic, err := loadEpic(dir, "default")
	if err != nil {
		t.Fatalf("Failed to load migrated epic: %v", err)
	}
	if len(epic.Milestones) != 1 {
		t.Errorf("len(Milestones) = %d, want 1", len(epic.Milestones))
	}

	// Check backup was created
	if _, err := os.Stat(configPath + ".bak"); os.IsNotExist(err) {
		t.Error("Backup file not created")
	}
}

// ============================================
// Artifact Tests
// ============================================

func TestArtifactExists(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	// Should not exist initially
	if artifactExists(dir, "test_artifact") {
		t.Error("Artifact should not exist initially")
	}

	// Create artifact
	writeFile(t, filepath.Join(dir, ".accelyst", "artifacts", "test_artifact.md"), "content")

	// Should exist now
	if !artifactExists(dir, "test_artifact") {
		t.Error("Artifact should exist after creation")
	}
}

func TestLoadArtifact(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	content := "# Test Artifact\n\nThis is test content."
	writeFile(t, filepath.Join(dir, ".accelyst", "artifacts", "test.md"), content)

	loaded, err := loadArtifact(dir, "test")
	if err != nil {
		t.Fatalf("loadArtifact failed: %v", err)
	}

	if loaded != content {
		t.Errorf("Content mismatch: got %q, want %q", loaded, content)
	}
}

func TestLoadArtifact_NotFound(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	// Create artifacts directory but not the file
	os.MkdirAll(filepath.Join(dir, ".accelyst", "artifacts"), 0755)

	_, err := loadArtifact(dir, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent artifact, got nil")
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

// ============================================
// Modifies Field Tests
// ============================================

func TestLoadEpic_WithModifies(t *testing.T) {
	dir := setupTestDir(t)
	defer cleanupTestDir(t, dir)

	epic := `name: "Test Epic"
milestones:
  - id: m1
    name: "Milestone One"
    steps:
      - id: impl_feature
        instruction: "implement feature"
        modifies:
          - client/app.js
          - client/index.html
`
	writeFile(t, filepath.Join(dir, ".accelyst", "epics", "test.yaml"), epic)

	e, err := loadEpic(dir, "test")
	if err != nil {
		t.Fatalf("loadEpic failed: %v", err)
	}

	step := e.Milestones[0].Steps[0]
	if len(step.Modifies) != 2 {
		t.Errorf("len(Modifies) = %d, want 2", len(step.Modifies))
	}
	if step.Modifies[0] != "client/app.js" {
		t.Errorf("Modifies[0] = %q, want %q", step.Modifies[0], "client/app.js")
	}
	if step.Modifies[1] != "client/index.html" {
		t.Errorf("Modifies[1] = %q, want %q", step.Modifies[1], "client/index.html")
	}
}

func TestGeneratePromptFragment_WithModifies(t *testing.T) {
	step := Step{
		ID:          "impl_download",
		Instruction: "implement download button",
		Modifies:    []string{"client/viewer.js", "client/index.html"},
	}

	result := generatePromptFragment(step, nil, "/tmp")

	if !strings.Contains(result, "[modifies: client/viewer.js, client/index.html]") {
		t.Errorf("Expected modifies clause in output, got: %s", result)
	}
}

func TestGeneratePromptFragment_NoModifies(t *testing.T) {
	step := Step{
		ID:          "analyze",
		Instruction: "analyze component",
	}

	result := generatePromptFragment(step, nil, "/tmp")

	if strings.Contains(result, "[modifies:") {
		t.Errorf("Step without modifies should not have modifies clause, got: %s", result)
	}
}

func TestGeneratePromptFragment_EmptyModifies(t *testing.T) {
	step := Step{
		ID:          "analyze",
		Instruction: "analyze component",
		Modifies:    []string{},
	}

	result := generatePromptFragment(step, nil, "/tmp")

	if strings.Contains(result, "[modifies:") {
		t.Errorf("Step with empty modifies should not have modifies clause, got: %s", result)
	}
}

func TestGeneratePromptFragment_FullStep(t *testing.T) {
	step := Step{
		ID:           "impl_feature",
		Instruction:  "implement the feature",
		ProjectParts: []string{"client"},
		DependsOn:    []string{"analyze_step"},
		Modifies:     []string{"client/app.js"},
	}

	projectParts := map[string]ProjectPart{
		"client": {Name: "client", DirectoryAbs: "/path/to/client"},
	}

	result := generatePromptFragment(step, projectParts, "/tmp")

	// Should have project parts
	if !strings.Contains(result, "Analyze client") {
		t.Errorf("Expected project parts, got: %s", result)
	}
	// Should have dependencies
	if !strings.Contains(result, "read results from steps analyze_step") {
		t.Errorf("Expected dependencies, got: %s", result)
	}
	// Should have modifies at the end
	if !strings.Contains(result, "[modifies: client/app.js]") {
		t.Errorf("Expected modifies clause, got: %s", result)
	}
	// Should NOT have subagent prefix (has dependencies)
	if strings.Contains(result, "use a subagent to") {
		t.Errorf("Step with dependencies should not have subagent prefix, got: %s", result)
	}
}

func TestProcessMilestone_WithModifies(t *testing.T) {
	milestone := Milestone{
		ID:   "test",
		Name: "Test",
		Steps: []Step{
			{ID: "analyze", Instruction: "analyze"},
			{
				ID:          "impl",
				Instruction: "implement",
				DependsOn:   []string{"analyze"},
				Modifies:    []string{"app.js", "config.js"},
			},
		},
	}

	result, err := processMilestone(milestone, nil, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Analyze step should not have modifies
	if strings.Contains(result, "analyze: use a subagent to analyze [modifies:") {
		t.Errorf("Analyze step should not have modifies clause")
	}

	// Impl step should have modifies
	if !strings.Contains(result, "[modifies: app.js, config.js]") {
		t.Errorf("Impl step should have modifies clause, got: %s", result)
	}
}
