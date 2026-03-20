package promptcatalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultCatalogLoadsBuiltinPrompts(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	spec, err := cat.Get("humanizer", "base")
	if err != nil {
		t.Fatalf("Get(humanizer, base) error = %v", err)
	}
	if spec.Kind != "humanizer" || spec.Name != "base" {
		t.Fatalf("unexpected spec: %#v", spec)
	}
}

func TestCatalogRenderReplacesVariables(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	rendered, spec, err := cat.Render("image", "cover-default", map[string]string{
		"ARTICLE_TITLE":   "测试标题",
		"ARTICLE_SUMMARY": "测试摘要",
		"VISUAL_STYLE":    "极简",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if spec.Name != "cover-default" {
		t.Fatalf("spec.Name = %q", spec.Name)
	}
	if !strings.Contains(rendered, "测试标题") || !strings.Contains(rendered, "极简") {
		t.Fatalf("rendered prompt = %q", rendered)
	}
}

func TestCatalogPrefersExplicitPromptDirOverBuiltin(t *testing.T) {
	ResetDefaultCatalogForTests()
	tmpDir := t.TempDir()
	overrideDir := filepath.Join(tmpDir, "prompts", "humanizer")
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	override := strings.Join([]string{
		"name: medium",
		"kind: humanizer",
		"description: override",
		"version: \"1.0\"",
		"template: |",
		"  override medium",
	}, "\n")
	if err := os.WriteFile(filepath.Join(overrideDir, "medium.yaml"), []byte(override), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv(promptsDirEnvVar, filepath.Join(tmpDir, "prompts"))

	cat := NewCatalog()
	if err := cat.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	spec, err := cat.Get("humanizer", "medium")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if spec.Source != filepath.Join(tmpDir, "prompts") {
		t.Fatalf("Source = %q", spec.Source)
	}
	if strings.TrimSpace(spec.Template) != "override medium" {
		t.Fatalf("Template = %q", spec.Template)
	}
}

func TestCatalogListFilteredByArchetypeAndTag(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	prompts := cat.ListFiltered(ListFilter{
		Kind:      "image",
		Archetype: "cover",
		Tag:       "editorial",
	})
	if len(prompts) == 0 {
		t.Fatal("expected filtered image prompts")
	}
	for _, prompt := range prompts {
		if prompt.Kind != "image" {
			t.Fatalf("unexpected kind: %#v", prompt)
		}
		if prompt.Archetype != "cover" {
			t.Fatalf("unexpected archetype: %#v", prompt)
		}
		if !containsTag(prompt.Tags, "editorial") {
			t.Fatalf("expected editorial tag: %#v", prompt)
		}
	}
}

func TestBuiltinImagePromptIncludesArchetypeMetadata(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	spec, err := cat.Get("image", "cover-hero")
	if err != nil {
		t.Fatalf("Get(image, cover-hero) error = %v", err)
	}
	if spec.Archetype != "cover" {
		t.Fatalf("Archetype = %q", spec.Archetype)
	}
	if !containsTag(spec.Tags, "hero") {
		t.Fatalf("Tags = %#v", spec.Tags)
	}
	if len(spec.Examples) == 0 {
		t.Fatalf("Examples = %#v", spec.Examples)
	}
	if spec.Metadata["author"] != "geekjourneyx" {
		t.Fatalf("Metadata = %#v", spec.Metadata)
	}
}

func TestBuiltinImagePromptSupportsAttributionMetadata(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	spec, err := cat.Get("image", "infographic-flat-vector-panorama")
	if err != nil {
		t.Fatalf("Get(image, infographic-flat-vector-panorama) error = %v", err)
	}
	if spec.Archetype != "infographic" {
		t.Fatalf("Archetype = %q", spec.Archetype)
	}
	if spec.Metadata["author"] != "geekjourneyx" {
		t.Fatalf("author metadata = %#v", spec.Metadata)
	}
	if spec.Metadata["inspired_by"] != "op7418" {
		t.Fatalf("inspired_by metadata = %#v", spec.Metadata)
	}
	if spec.Metadata["provenance"] != "adapted" {
		t.Fatalf("provenance metadata = %#v", spec.Metadata)
	}
}
