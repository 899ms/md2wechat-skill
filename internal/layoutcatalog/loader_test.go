package layoutcatalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBuiltinIncludesHero(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	spec, ok := c.Get("hero")
	if !ok {
		t.Fatalf("expected hero module to be present")
	}
	if spec.Category != "opening" {
		t.Errorf("hero.Category = %q, want opening", spec.Category)
	}
}

func TestEnvOverrideTrumpsBuiltin(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Cleanup(ResetDefaultCatalogForTests)

	dir := t.TempDir()
	override := filepath.Join(dir, "opening")
	if err := os.MkdirAll(override, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := []byte(`schema_version: "1"
name: hero
version: "999.0.0"
category: opening
serves: [attention]
metadata:
  author: test
  provenance: override
`)
	if err := os.WriteFile(filepath.Join(override, "hero.yaml"), yaml, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MD2WECHAT_LAYOUT_DIR", dir)

	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	spec, _ := c.Get("hero")
	if spec.Version != "999.0.0" {
		t.Errorf("override not applied: version = %q", spec.Version)
	}
}

func TestParseLayoutSpecRejectsInvalidServes(t *testing.T) {
	yaml := []byte(`schema_version: "1"
name: bad
version: "1.0.0"
category: opening
serves: [bogus]
metadata:
  author: x
  provenance: builtin
`)
	_, err := parseLayoutSpec(yaml)
	if err == nil {
		t.Fatal("expected error for invalid serves value")
	}
}
