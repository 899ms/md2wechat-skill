package layoutcatalog

import (
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPinnedBrandSymbolMotionSupportMatchesDiscovery(t *testing.T) {
	data, err := os.ReadFile("testdata/upstream_brand_symbol_motions.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		SourceCommit string              `yaml:"source_commit"`
		SourceSHA256 string              `yaml:"source_sha256"`
		Symbols      map[string][]string `yaml:"symbols"`
	}
	if err := yaml.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	if source.SourceCommit != "2795cf552383fa6883733b7dec56dc7b3bda8aa1" || source.SourceSHA256 != "25c5654a295e31b554588ba909cb538781421aebaa713f406f8558f4727c5759" || len(source.Symbols) != 12 {
		t.Fatalf("symbol source drift: %+v", source)
	}
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"hero", "section-title", "closing", "author-card"} {
		spec, ok := c.Get(name)
		if !ok {
			t.Fatal(name)
		}
		var symbolField, motionField FieldSpec
		for _, field := range spec.Fields.Optional {
			if field.Name == "symbol" {
				symbolField = field
			}
			if field.Name == "motion" {
				motionField = field
			}
		}
		for symbol := range source.Symbols {
			if !slices.Contains(symbolField.Enum, symbol) {
				t.Errorf("%s omits symbol %s", name, symbol)
			}
		}
		for motion, symbols := range motionField.SymbolKeysByValue {
			var expected []string
			for _, symbol := range symbolField.Enum {
				if slices.Contains(source.Symbols[symbol], motion) {
					expected = append(expected, symbol)
				}
			}
			if !slices.Equal(symbols, expected) {
				t.Errorf("%s motion %s supports %v, want %v", name, motion, symbols, expected)
			}
		}
		for _, motion := range []string{"draw", "draw-stagger", "draw-reverse", "scale-in", "rotate-in", "draw-fill"} {
			if _, ok := motionField.SymbolKeysByValue[motion]; !ok {
				t.Errorf("%s omits symbol support for %s", name, motion)
			}
		}
	}
}

func TestMilestoneMotionPlacementMatrix(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, markdown string
		valid          bool
	}{
		{"hero seal stamp", ":::hero\nvariant: seal\ntitle: Launch\nsymbol: rounded-seal\nmotion: stamp-in\n:::", true},
		{"hero journal stamp", ":::hero\nvariant: journal\ntitle: Launch\nsymbol: rounded-seal\nmotion: stamp-in\n:::", false},
		{"section divider center", ":::section-title\nvariant: divider\ntitle: Part\nmotion: draw-center\n:::", true},
		{"section marker center", ":::section-title\nvariant: marker\ntitle: Part\nmotion: draw-center\n:::", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := c.Validate(tt.markdown)
			if tt.valid && len(report.Errors) != 0 {
				t.Fatalf("unexpected errors: %+v", report.Errors)
			}
			if !tt.valid && len(report.Errors) == 0 {
				t.Fatal("unsupported placement accepted")
			}
		})
	}
}

func TestRecommendedGalleryAcceptsOldSyntaxAndRequiresImage(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	spec, ok := c.Get("gallery")
	if !ok || spec.Lifecycle != LifecycleRecommended {
		t.Fatalf("gallery lifecycle: %+v", spec)
	}
	for _, markdown := range []string{
		":::gallery[主题画廊]\n![一](https://example.com/1.png)\n:::",
		":::gallery\n![一](https://example.com/1.png)\n![二](https://example.com/2.png)\n:::",
	} {
		if report := c.Validate(markdown); len(report.Errors) != 0 {
			t.Fatalf("%q: %+v", markdown, report.Errors)
		}
	}
	if report := c.Validate(":::gallery\n无图片\n:::"); len(report.Errors) == 0 {
		t.Fatal("zero image accepted")
	}
}

func TestMilestoneSymbolMotionSupportMatrix(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, markdown string
		valid          bool
	}{
		{"hero supported", ":::hero\nvariant: seal\ntitle: Launch\nsymbol: rounded-seal\nmotion: stamp-in\n:::", true},
		{"hero unsupported symbol", ":::hero\nvariant: seal\ntitle: Launch\nsymbol: lens\nmotion: stamp-in\n:::", false},
		{"masthead classic symbol", ":::hero\nvariant: masthead\ntitle: Launch\nsymbol: spark-solid\nmotion: draw\n:::", false},
		{"masthead brand symbol", ":::hero\nvariant: masthead\ntitle: Launch\nsymbol: mountain\nmotion: draw\n:::", true},
		{"journal classic symbol falls back", ":::hero\nvariant: journal\ntitle: Launch\nsymbol: circle\n:::", false},
		{"seal classic symbol falls back", ":::hero\nvariant: seal\ntitle: Launch\nsymbol: star\n:::", false},
		{"orbit classic symbol falls back", ":::hero\nvariant: orbit\ntitle: Launch\nsymbol: infinity\n:::", false},
		{"journal brand symbol", ":::hero\nvariant: journal\ntitle: Launch\nsymbol: mountain\n:::", true},
		{"closing unsupported fill", ":::closing\ntitle: End\nsymbol: lens\nmotion: draw-fill\n:::", false},
		{"closing supported fill", ":::closing\ntitle: End\nsymbol: nested-diamonds\nmotion: draw-fill\n:::", true},
		{"author motion without symbol", ":::author-card\nname: Agent\nmotion: draw\n:::", false},
		{"hero title effect short", ":::hero\nvariant: journal\ntitle: Brand\nmotion: focus-in\n:::", true},
		{"hero title effect long", ":::hero\nvariant: journal\ntitle: This title is too long\nmotion: focus-in\n:::", false},
		{"hero title effect markdown", ":::hero\nvariant: journal\ntitle: **Bold**\nmotion: wipe-in\n:::", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			report := c.Validate(tt.markdown)
			if tt.valid && len(report.Errors) != 0 {
				t.Fatalf("unexpected errors: %+v", report.Errors)
			}
			if !tt.valid && len(report.Errors) == 0 {
				t.Fatal("ineffective motion accepted")
			}
		})
	}
}
