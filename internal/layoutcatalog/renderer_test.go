package layoutcatalog

import (
	"strings"
	"testing"
)

func TestRenderHeroFields(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	out, err := c.Render("hero", map[string]any{
		"eyebrow":  "深度观察",
		"title":    "公众号排版的真问题不是好不好看",
		"subtitle": "是读者愿不愿意读完",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.HasPrefix(out, ":::hero") || !strings.HasSuffix(strings.TrimRight(out, "\n"), ":::") {
		t.Errorf("output missing :::hero block fence:\n%s", out)
	}
	for _, want := range []string{"eyebrow: 深度观察", "title: 公众号排版的真问题不是好不好看", "subtitle: 是读者愿不愿意读完"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestRenderMissingRequiredFieldFails(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	_, err := c.Render("hero", map[string]any{"eyebrow": "x"})
	if err == nil {
		t.Fatal("expected error for missing title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention missing field name, got: %v", err)
	}
}

func TestRenderUnknownModuleFails(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	_, err := c.Render("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
