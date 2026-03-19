package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadStyleAppliesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	stylePath := filepath.Join(tmpDir, "style.yaml")
	content := strings.Join([]string{
		`english_name: custom-style`,
		`writing_prompt: Write with intent.`,
	}, "\n")
	if err := os.WriteFile(stylePath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	sm := NewStyleManager()
	if err := sm.loadStyle(stylePath); err != nil {
		t.Fatalf("loadStyle() error = %v", err)
	}

	style, ok := sm.styles["custom-style"]
	if !ok {
		t.Fatal("expected custom-style to be loaded")
	}
	if style.Name != "custom-style" {
		t.Fatalf("Name = %q", style.Name)
	}
	if style.Category != "自定义" {
		t.Fatalf("Category = %q", style.Category)
	}
	if style.Version != "1.0" {
		t.Fatalf("Version = %q", style.Version)
	}
}

func TestGetStyleWithPromptInterpolatesTemplateVariables(t *testing.T) {
	sm := &StyleManager{
		styles: map[string]*WriterStyle{
			DefaultStyleName: {
				Name:          "Dan Koe",
				EnglishName:   DefaultStyleName,
				WritingPrompt: "Title: {title}\nBody: {body}",
			},
		},
		initialized: true,
	}

	style, err := sm.GetStyleWithPrompt(DefaultStyleName, map[string]string{
		"title": "Test Title",
		"body":  "Test Body",
	})
	if err != nil {
		t.Fatalf("GetStyleWithPrompt() error = %v", err)
	}
	if !strings.Contains(style.WritingPrompt, "Test Title") || !strings.Contains(style.WritingPrompt, "Test Body") {
		t.Fatalf("WritingPrompt = %q", style.WritingPrompt)
	}
	if sm.styles[DefaultStyleName].WritingPrompt != "Title: {title}\nBody: {body}" {
		t.Fatalf("original style was mutated: %q", sm.styles[DefaultStyleName].WritingPrompt)
	}
}

func TestValidateStyleRequiresCoreFields(t *testing.T) {
	sm := NewStyleManager()

	if err := sm.ValidateStyle(&WriterStyle{WritingPrompt: "prompt"}); err == nil {
		t.Fatal("expected english_name validation error")
	}
	if err := sm.ValidateStyle(&WriterStyle{EnglishName: "custom"}); err == nil {
		t.Fatal("expected writing_prompt validation error")
	}
}

func TestGetWritersDirPrefersExplicitEnvironmentVariable(t *testing.T) {
	customDir := filepath.Join(t.TempDir(), "custom-writers")
	t.Setenv(writersDirEnvVar, customDir)

	sm := NewStyleManager()
	if got := sm.getWritersDir(); got != customDir {
		t.Fatalf("getWritersDir() = %q, want %q", got, customDir)
	}
}
